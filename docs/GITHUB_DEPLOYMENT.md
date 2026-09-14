# GitHub Actions deployment

The public alpha is deployed only from the protected `deploy` branch. The
`production` GitHub Environment must have required reviewers enabled. Normal
development pushes and pull requests never activate the VPS. The workflow also
checks the exact Git ref, so a manual run must explicitly select `deploy`.

The workflow builds and tests on an ephemeral GitHub runner, prepares a content
addressed release, audits the exact image IDs, transfers the archive over SSH,
and runs the repository's host preflight and immutable activation checks. It does
not use a container registry. The VPS receives the release and audit evidence,
not the Git checkout. A deployment is rejected if the release revision, image
identities, audit, host controls, or tunnel credential do not match.

## One-time values

Create these repository or `production` Environment secrets in GitHub Settings →
Secrets and variables → Actions:

| Secret | Where it comes from |
| --- | --- |
| `VPS_HOST` | The VPS IPv4 address from OVHcloud, or a stable DNS/SSH name. |
| `VPS_USER` | The non-password SSH user configured on the VPS (`root`, `ubuntu`, or `debian` depending on the image). Prefer a dedicated deployment user in the `docker` group. |
| `VPS_SSH_PRIVATE_KEY` | A dedicated deployment key whose public key is installed in the VPS user's `~/.ssh/authorized_keys`. Keep the private key only in GitHub's encrypted secret. |
| `VPS_KNOWN_HOSTS` | The verified output of `ssh-keyscan -H VPS_HOST`; compare the host-key fingerprint with the OVH console before saving it. |
| `CLOUDFLARE_TUNNEL_TOKEN` | Cloudflare Zero Trust → Networks → Tunnels → the named tunnel → Docker installation token. The workflow writes it to the VPS as a mode-0600 file and never prints it. |

The Cloudflare tunnel itself is created once in the Cloudflare dashboard. Add a
public hostname for `jamgati.com` pointing to `http://web:8080`, and configure a
final unmatched-hostname rule that returns 404. Keep the VPS web/API/Valkey ports
closed to the Internet. The token permits the `cloudflared` process on this VPS to
join that tunnel; it is not an application or participant credential.

The workflow generates `vapid.json` on the VPS on its first deployment and retains
it across releases. It does not regenerate or export the private key. Operational
email alerts are separate: install `/etc/gati/alerts.env` on the VPS from
`deploy/alerts.env.example` with the SMTP provider's host, TLS mode, username, app
password, sender and project recipient. The application can run without that
optional alert service.

## First VPS preparation

Copy and run [`scripts/vps-bootstrap.sh`](../scripts/vps-bootstrap.sh) from the
OVH console or an initial SSH session as root. It is idempotent and prompts for
the public key that GitHub Actions will use; passing `GATI_SSH_PUBLIC_KEY` in the
environment is also supported. It does not ask for or store the private key.

```sh
scp scripts/vps-bootstrap.sh root@VPS_IP:/root/
ssh root@VPS_IP 'bash /root/vps-bootstrap.sh'
```

Install Docker Engine and grant the deployment user access to Docker. Configure the
host firewall so only SSH and the management path are available; the application
ports remain private because Cloudflare Tunnel connects outbound. Disable swap,
hibernation, core dumps and provider RAM snapshots, and ensure `/opt/gati` is
writable by the deployment user. The workflow's `scripts/release.py preflight`
checks the settings it can observe and refuses public activation otherwise.

After the first successful run, inspect the release manifest, audit JSON, running
image IDs, container mounts and Cloudflare response headers independently. Keep the
previous release directory for rollback. The workflow intentionally does not delete
old releases or participant state.

## Triggering and rollback

Merge reviewed code into `deploy`, then approve the `production` Environment gate.
The workflow serializes deployments so two releases cannot activate concurrently.
The `deploy` branch should be protected against direct unreviewed pushes.

Rollback remains an explicit operator action using the existing command, after
checking configuration compatibility:

```sh
python3 scripts/release.py rollback \
  --release /opt/gati/releases/PREVIOUS_SHA \
  --current /opt/gati/releases/CURRENT_SHA \
  --edge --audit /opt/gati/audits/PREVIOUS_SHA/audit.json
```

GitHub-hosted runners can see repository contents, deployment timing and the
temporary use of encrypted secrets. OVHcloud and Cloudflare retain their own
network/account metadata. This automation does not remove those provider
visibility boundaries.
