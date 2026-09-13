package notification

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/ecdh"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"math"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	webpush "github.com/SherClockHolmes/webpush-go"
	"github.com/jegati/jegati/internal/config"
	"github.com/jegati/jegati/internal/store"
)

type Keys struct {
	Public  string `json:"public_key"`
	Private string `json:"private_key"`
}
type Registration struct {
	Revision  int64  `json:"revision"`
	Binding   string `json:"binding"`
	Endpoint  string `json:"endpoint"`
	P256DH    string `json:"p256dh"`
	Auth      string `json:"auth"`
	ExpiresAt int64  `json:"expires_at"`
}
type Service struct {
	Store  *store.Store
	config config.Config
	keys   Keys
	aead   cipher.AEAD
	client webpush.HTTPClient
}

func GenerateKeys() (Keys, error) {
	private, public, e := webpush.GenerateVAPIDKeys()
	return Keys{public, private}, e
}
func New(c config.Config, s *store.Store, path string) (*Service, error) {
	if !c.Notifications.PushEnabled {
		return nil, nil
	}
	if s == nil || c.Profile == "simulation" {
		return nil, errors.New("push needs a live production-profile store; simulation delivery is forbidden")
	}
	file, e := os.Open(path)
	if e != nil {
		return nil, errors.New("push key file required")
	}
	defer file.Close()
	var keys Keys
	decoder := json.NewDecoder(io.LimitReader(file, 2049))
	decoder.DisallowUnknownFields()
	if decoder.Decode(&keys) != nil || decoder.Decode(new(any)) != io.EOF {
		return nil, errors.New("invalid push key file")
	}
	return newService(c, s, keys, protectedClient(c.Notifications.PushEndpointHosts, time.Duration(c.Notifications.PushTimeoutSeconds)*time.Second, c.Notifications.PushConcurrency))
}
func newService(c config.Config, s *store.Store, keys Keys, client webpush.HTTPClient) (*Service, error) {
	private, e := base64.RawURLEncoding.DecodeString(keys.Private)
	if e != nil {
		return nil, errors.New("invalid push key")
	}
	key, e := ecdh.P256().NewPrivateKey(private)
	if e != nil || base64.RawURLEncoding.EncodeToString(key.PublicKey().Bytes()) != keys.Public {
		return nil, errors.New("invalid push key pair")
	}
	derived := sha256.Sum256(append([]byte("GATI temporary push transport v1\x00"), private...))
	block, e := aes.NewCipher(derived[:])
	if e != nil {
		return nil, e
	}
	aead, e := cipher.NewGCM(block)
	if e != nil {
		return nil, e
	}
	return &Service{Store: s, config: c, keys: keys, aead: aead, client: client}, nil
}
func (s *Service) PublicKey() string { return s.keys.Public }
func (s *Service) Validate(r Registration) error {
	if !regexp.MustCompile(`^[a-f0-9]{64}$`).MatchString(r.Binding) || r.ExpiresAt <= 0 || r.Revision < 0 {
		return errors.New("invalid subscription")
	}
	if _, e := endpointURL(r.Endpoint, s.config.Notifications.PushEndpointHosts); e != nil {
		return e
	}
	raw, e := base64.RawURLEncoding.DecodeString(r.P256DH)
	if e != nil {
		return errors.New("invalid subscription key")
	}
	if _, e = ecdh.P256().NewPublicKey(raw); e != nil {
		return errors.New("invalid subscription key")
	}
	auth, e := base64.RawURLEncoding.DecodeString(r.Auth)
	if e != nil || len(auth) != 16 {
		return errors.New("invalid subscription authentication")
	}
	return nil
}
func (s *Service) Register(ctx context.Context, hash string, r Registration) (int64, error) {
	if e := s.Validate(r); e != nil {
		return 0, e
	}
	raw, _ := json.Marshal(r)
	nonce := make([]byte, s.aead.NonceSize())
	if _, e := rand.Read(nonce); e != nil {
		return 0, errors.New("randomness unavailable")
	}
	aad := []byte(hash + ":" + r.Binding)
	encrypted := s.aead.Seal(nonce, nonce, raw, aad)
	digest := sha256.Sum256(append(aad, raw...))
	return s.Store.RegisterPush(ctx, hash, store.PushBinding{Revision: r.Revision, Binding: r.Binding, Digest: hex.EncodeToString(digest[:]), Ciphertext: base64.RawURLEncoding.EncodeToString(encrypted), ExpiresAt: r.ExpiresAt})
}
func (s *Service) decrypt(hash string, job store.PushJob) (Registration, error) {
	raw, e := base64.RawURLEncoding.DecodeString(job.Ciphertext)
	if e != nil || len(raw) < s.aead.NonceSize()+s.aead.Overhead() || len(raw) > 4096 {
		return Registration{}, errors.New("invalid stored transport")
	}
	plain, e := s.aead.Open(nil, raw[:s.aead.NonceSize()], raw[s.aead.NonceSize():], []byte(hash+":"+job.Binding))
	if e != nil {
		return Registration{}, errors.New("invalid stored transport")
	}
	var r Registration
	if json.Unmarshal(plain, &r) != nil || s.Validate(r) != nil || r.Binding != job.Binding {
		return Registration{}, errors.New("invalid stored transport")
	}
	return r, nil
}
func (s *Service) policy() store.PushPolicy {
	n := s.config.Notifications
	return store.PushPolicy{PollMS: int64(n.ForegroundPollSeconds) * 1000, QueueMS: int64(n.QueueTtlSeconds) * 1000, GapMS: int64(max(n.PushMinIntervalSeconds, int(math.Ceil(3600/float64(n.PushMaxPerHour))))) * 1000, RetryMS: int64(n.PushRetrySeconds) * 1000, LeaseMS: int64(n.PushTimeoutSeconds+2) * 1000, MinimumMS: int64(s.config.Matching.LateJoinMinRemainingMinutes) * 60000, MaxAttempts: n.PushMaxAttempts}
}
func (s *Service) Run(ctx context.Context) {
	ticker := time.NewTicker(time.Duration(s.config.Notifications.PushWorkerSeconds) * time.Second)
	defer ticker.Stop()
	for {
		_ = s.Step(ctx)
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}
}
func (s *Service) Step(parent context.Context) error {
	ctx, cancel := context.WithTimeout(parent, time.Duration(s.config.Notifications.PushWorkerSeconds+s.config.Notifications.PushTimeoutSeconds)*time.Second)
	defer cancel()
	hashes, e := s.Store.PushDue(ctx, s.config.Notifications.PushWorkerBatchSize)
	if e != nil {
		return errors.New("notification scan unavailable")
	}
	work := make(chan string)
	var wg sync.WaitGroup
	for n := 0; n < min(len(hashes), s.config.Notifications.PushConcurrency); n++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for hash := range work {
				s.deliver(ctx, hash)
			}
		}()
	}
	for _, hash := range hashes {
		select {
		case work <- hash:
		case <-ctx.Done():
			close(work)
			wg.Wait()
			return ctx.Err()
		}
	}
	close(work)
	wg.Wait()
	return ctx.Err()
}
func (s *Service) deliver(ctx context.Context, hash string) {
	raw := make([]byte, 16)
	if _, e := rand.Read(raw); e != nil {
		return
	}
	policy := s.policy()
	job, e := s.Store.ClaimPush(ctx, hash, hex.EncodeToString(raw), policy)
	if e != nil || job == nil {
		return
	}
	outcome, poll := s.send(ctx, hash, *job, policy.PollMS)
	_ = s.Store.FinishPush(ctx, hash, job.Claim, outcome, poll)
}
func (s *Service) send(ctx context.Context, hash string, job store.PushJob, poll int64) (string, int64) {
	r, e := s.decrypt(hash, job)
	if e != nil {
		return "gone", poll
	}
	binding, e := s.Store.PushStatus(ctx, hash)
	if e == store.ErrGone {
		return "gone", poll
	}
	if e != nil {
		return "retry", poll
	}
	if binding.Binding != job.Binding {
		return "sent", poll
	}
	signal, e := s.Store.Status(ctx, hash)
	if e == store.ErrGone {
		return "gone", poll
	}
	if e != nil {
		return "retry", poll
	}
	now, e := s.Store.Now(ctx)
	if e != nil {
		return "retry", poll
	}
	until := min(job.Until, job.ExpiresAt, r.ExpiresAt, signal.ExpiresAt, signal.GatheringUntil)
	if signal.PushRevision != r.Revision || signal.Gathering == "" || until-now < 1000 {
		return "sent", poll
	}
	payload, _ := json.Marshal(struct {
		Binding   string `json:"binding"`
		ExpiresAt int64  `json:"expires_at"`
	}{job.Binding, until})
	subscription := webpush.Subscription{Endpoint: r.Endpoint, Keys: webpush.Keys{P256dh: r.P256DH, Auth: r.Auth}}
	contact := strings.TrimPrefix(s.config.Notifications.PushContact, "mailto:") // upstream accepts raw email or HTTPS
	response, e := webpush.SendNotificationWithContext(ctx, payload, &subscription, &webpush.Options{HTTPClient: s.client, RecordSize: 512, Subscriber: contact, VAPIDPublicKey: s.keys.Public, VAPIDPrivateKey: s.keys.Private, TTL: int((until - now) / 1000), Topic: "gati-current", Urgency: webpush.UrgencyNormal})
	if e != nil {
		return "retry", poll
	}
	defer response.Body.Close()
	_, _ = io.Copy(io.Discard, io.LimitReader(response.Body, int64(s.config.Limits.MaxBodyBytes)))
	if response.StatusCode >= 200 && response.StatusCode < 300 {
		return "sent", poll
	}
	if response.StatusCode == 404 || response.StatusCode == 410 || response.StatusCode == 400 || response.StatusCode == 401 || response.StatusCode == 403 {
		return "gone", poll
	}
	if seconds, e := strconv.Atoi(response.Header.Get("Retry-After")); e == nil && seconds > 0 {
		poll = max(poll, min(int64(seconds), int64(s.config.Notifications.QueueTtlSeconds))*1000)
	} else if date, e := http.ParseTime(response.Header.Get("Retry-After")); e == nil {
		poll = max(poll, min(date.Sub(time.Now()).Milliseconds(), int64(s.config.Notifications.QueueTtlSeconds)*1000))
	}
	return "retry", poll
}
