import { test as base, expect } from '@playwright/test';
// Playwright Firefox's override deliberately sets timestamp one day ahead:
// https://github.com/microsoft/playwright/blob/main/browser_patches/firefox/juggler/content/main.js
// Correct only known emulation values; production freshness checks stay intact.
export const test=base.extend<{ browserLocationClock: void }>({
 browserLocationClock:[async({context,browserName},use)=>{
  if(browserName==='firefox'||browserName==='webkit')await context.addInitScript(engine=>{
   const native=navigator.geolocation.getCurrentPosition.bind(navigator.geolocation);
   navigator.geolocation.getCurrentPosition=(success,error,options)=>native(position=>{
    const now=Date.now(),offset=position.timestamp-now;
    const firefox=engine==='firefox'&&offset>23*3600000&&offset<=25*3600000;
    // This pinned WebKit override returns microseconds; retain its actual age.
    const webkit=engine==='webkit'&&Math.abs(position.timestamp/1000-now)<3600000;
    if(firefox||webkit){
     const fixed={coords:position.coords,timestamp:firefox?now:position.timestamp/1000,toJSON:()=>({})} as GeolocationPosition;
     success(fixed);
    }else success(position);
   },error,options);
  },browserName);
  await use();
 },{auto:true}],
});
export {expect};
