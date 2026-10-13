// Offline Chrome, synthetic backup metadata only; no local session, secrets or mutations.
import assert from 'node:assert/strict';
import fs from 'node:fs/promises';
import path from 'node:path';
import {createRequire} from 'node:module';
import test,{after} from 'node:test';
const require=createRequire(import.meta.url);
const modules=process.env.LOCALADMIN_NODE_MODULES||path.join(process.env.USERPROFILE,'.cache/codex-runtimes/codex-primary-runtime/dependencies/node/node_modules');
const {chromium}=require(path.join(modules,'playwright'));
const browser=await chromium.launch({channel:'chrome',headless:true});after(()=>browser.close());
const html=(await fs.readFile('internal/localadmin/assets/index.html','utf8')).replace(/<script\b[^>]*>[\s\S]*?<\/script>/g,'').replace(/<link\b[^>]*>/g,'');
const app=await fs.readFile('internal/localadmin/assets/app.js','utf8'),backups=await fs.readFile('internal/localadmin/assets/backups.js','utf8');
const entry=i=>({id:`20261011T000000.000000000Z-${i.toString(16).padStart(16,'0')}${i<45?'-control':'.json'}`,scope:i<45?'control-db':i<55?'startup-config':'admin-options',valid:true,integrity:i<45,version:1,size:512+i,createdAt:'2026-10-11T00:00:00Z',config:'fixture-private-config',token:'fixture-private-ticket',SHA256:'fixture-private-digest'});
const catalog={backups:Array.from({length:65},(_,i)=>entry(i)),scopes:{'control-db':true,'startup-config':true,'admin-options':true},more:true,secret:'fixture-private-catalog'};
async function fixture(t,data=catalog,timeZone="Asia/Shanghai"){const context=await browser.newContext({acceptDownloads:true,timezoneId:timeZone});t.after(()=>context.close());await context.route('**/*',r=>r.abort());const p=await context.newPage();await p.setContent(html);await p.addScriptTag({content:app});await p.addScriptTag({content:backups});await p.evaluate(data=>{authenticated=true;csrf='catalog-session';window.showOverview=()=>{};window.showLogs=()=>{};document.getElementById('auth').hidden=true;document.getElementById('workspace').hidden=false;document.getElementById('overview-view').hidden=true;document.getElementById('backups-view').hidden=false;window.data=data;window.requests=[];window.pending=[];window.mode='success';window.revoked=[];const revoke=URL.revokeObjectURL.bind(URL);URL.revokeObjectURL=u=>{revoked.push(u);revoke(u)};window.fetch=async(url,options)=>{requests.push({url,method:options.method||'GET'});if(mode==='pending')return new Promise(resolve=>pending.push(resolve));return {ok:mode==='success',status:mode==='success'?200:503,json:async()=>mode==='success'?window.data:{error:'fixture read failed'}}};},data);return p;}
const idle=p=>p.waitForFunction(()=>!backupBusy);
const read=async p=>{await p.evaluate(()=>showBackups(true));await idle(p);};
async function download(p){const [d]=await Promise.all([p.waitForEvent('download'),p.locator('#download-backup-catalog').click()]);const text=await fs.readFile(await d.path(),'utf8');return {d,text,value:JSON.parse(text)};}

const ids=p=>p.locator('#backup-list .backup-card small').allTextContents();
const dated=(i,time,overrides={})=>({...entry(i),createdAt:time,...overrides});
const dates={...catalog,backups:[
 dated(0,'2026-10-10T15:59:59.999Z'),dated(1,'2026-10-10T16:00:00Z'),
 ...Array.from({length:55},(_,i)=>dated(i+2,'2026-10-11T03:00:00Z')),
 dated(57,'2026-10-11T15:59:59.999Z'),dated(58,'2026-10-11T16:00:00Z'),
 dated(59,'bad-date'),dated(60,'2026-10-11T03:00:00Z',{valid:false}),dated(61,null)
]};
async function apply(p,from='',until=''){await p.locator('#backup-from-date').fill(from);await p.locator('#backup-until-date').fill(until);await p.locator('#apply-backup-date').click();}

test('local day includes both edges, filters all pages before sorting, and updates size totals without requests',async t=>{
 const p=await fixture(t,dates);await read(p);await p.locator('#backup-next').click();await apply(p,'2026-10-11','2026-10-11');
 assert.equal(await p.evaluate(()=>backupPage),1);assert.equal(await p.evaluate(()=>backupMatchCount),57);assert.match(await p.locator('#backup-size-matched').textContent(),/57 份/);
 await p.selectOption('#backup-sort','oldest');const all=[];for(let i=0;i<3;i++){all.push(...await ids(p));if(i<2)await p.locator('#backup-next').click()};assert.deepEqual(all,dates.backups.slice(1,58).map(e=>e.id));assert.equal(new Set(all).size,57);
 assert.deepEqual(await p.evaluate(()=>[backupDateStart,backupDateEnd]),[Date.parse('2026-10-10T16:00:00Z'),Date.parse('2026-10-11T16:00:00Z')]);
 await p.selectOption('#backup-scope-filter','control-db');assert.equal(await p.evaluate(()=>backupMatchCount),44);await p.selectOption('#backup-state-filter','legacy');assert.equal(await p.evaluate(()=>backupMatchCount),0);
 await p.selectOption('#backup-state-filter','verified');await p.locator('#backup-search').fill(entry(1).id);assert.deepEqual(await ids(p),[entry(1).id]);assert.equal(await p.evaluate(()=>requests.length),1);
});

test('unapplied inputs and reversed dates preserve results; single bounds exclude unknown dates and clear restores all records',async t=>{
 const p=await fixture(t,dates);await read(p);await apply(p,'2026-10-11','2026-10-11');const previous=await ids(p);
 await p.locator('#backup-from-date').fill('2026-10-12');assert.match(await p.locator('#backup-date-status').textContent(),/尚未应用/);assert.deepEqual(await ids(p),previous);await p.locator('#apply-backup-date').click();assert.match(await p.locator('#backup-date-error').textContent(),/不能晚于/);assert.deepEqual(await ids(p),previous);assert.equal(await p.evaluate(()=>backupFromDate),'2026-10-11');
 const invalid=await p.evaluate(()=>['2026-02-30','2025-02-29','0000-01-01','invalid'].every(value=>{try{backupDateBoundary(value);return false}catch{return true}}));assert.equal(invalid,true);
 await apply(p,'','2026-10-10');assert.deepEqual(await ids(p),[entry(0).id]);await apply(p,'2026-10-12','');assert.deepEqual(await ids(p),[entry(58).id]);
 await p.selectOption('#backup-sort','largest');await p.locator('#clear-backup-filter').click();assert.equal(await p.evaluate(()=>backupMatchCount),dates.backups.length);assert.equal(await p.locator('#backup-from-date').inputValue(),'');assert.equal(await p.locator('#backup-until-date').inputValue(),'');assert.equal(await p.locator('#backup-date-error').isHidden(),true);assert.equal(await p.locator('#backup-sort').inputValue(),'largest');assert.equal(await p.evaluate(()=>backupDateStart===null&&backupDateEnd===null),true);
 await p.selectOption('#backup-state-filter','invalid');assert.deepEqual(await ids(p),[entry(60).id]);assert.equal(await p.evaluate(()=>requests.length),1);
});

test('catalog exports applied dates and browser timezone across all matched pages, preserving restore and cleanup targets',async t=>{
 const p=await fixture(t,dates);await read(p);await p.evaluate(()=>{backupPreview={backup:data.backups[0],token:'fixture-private-ticket',changed:true,expiresAt:Math.floor(Date.now()/1000)+600};backupCleanupPreview={token:'fixture-private-cleanup',candidates:[{id:'fixture'}],expiresAt:Math.floor(Date.now()/1000)+600};document.getElementById('restore-backup-id').textContent=data.backups[0].id;document.getElementById('restore-confirm').checked=true;document.getElementById('cleanup-confirm').checked=true;renderBackupExpiry()});
 await apply(p,'2026-10-11','2026-10-11');await p.selectOption('#backup-sort','smallest');await p.locator('#backup-next').click();await p.locator('#backup-from-date').fill('2026-10-12');const out=await download(p);
 assert.deepEqual(out.value.filters,{scope:'all',state:'all',keywordApplied:false,sort:'smallest',fromDate:'2026-10-11',untilDate:'2026-10-11',dateTimeZone:'Asia/Shanghai'});assert.equal(out.value.backups.length,57);assert.deepEqual(out.value.backups.map(e=>e.id),dates.backups.slice(1,58).map(e=>e.id));assert.doesNotMatch(out.text,/fixture-private|"(?:config|token|SHA256)"\s*:/);
 assert.equal(await p.evaluate(()=>backupPage),2);assert.equal(await p.locator('#restore-confirm').isChecked(),true);assert.equal(await p.locator('#cleanup-confirm').isChecked(),true);assert.equal(await p.locator('#restore-backup-id').textContent(),entry(0).id);assert.equal(await p.evaluate(()=>backupPreview.token),'fixture-private-ticket');assert.equal(await p.evaluate(()=>backupCleanupPreview.token),'fixture-private-cleanup');assert.equal(await p.evaluate(()=>requests.length),1);
 await p.locator('#clear-backup-filter').click();const clear=await download(p);assert.equal(Object.hasOwn(clear.value.filters,'fromDate'),false);
});

test('unread and busy controls block applying; refresh failure and navigation retain applied dates and draft',async t=>{
 const p=await fixture(t,dates);assert.equal(await p.locator('#apply-backup-date').isDisabled(),true);await p.evaluate(()=>{backupActive=true;applyBackupDateFilter()});assert.equal(await p.evaluate(()=>backupDateStart),null);await read(p);await apply(p,'2026-10-11','2026-10-11');await p.locator('#backup-until-date').fill('2026-10-12');
 await p.evaluate(()=>mode='pending');await p.locator('#reload-backups').click();await p.waitForFunction(()=>pending.length===1);assert.equal(await p.locator('#backup-from-date').isDisabled(),true);await p.evaluate(()=>applyBackupDateFilter());assert.equal(await p.evaluate(()=>backupUntilDate),'2026-10-11');
 await p.evaluate(()=>pending[0]({ok:false,status:503,json:async()=>({error:'fixture failure'})}));await idle(p);assert.equal(await p.evaluate(()=>backupMatchCount),57);assert.match(await p.locator('#backup-filter-count').textContent(),/刷新失败/);assert.equal(await p.locator('#backup-until-date').inputValue(),'2026-10-12');assert.match(await p.locator('#backup-date-status').textContent(),/尚未应用/);
 await p.evaluate(()=>{showBackups(false);applyBackupDateFilter();mode='success';showBackups(true)});await idle(p);assert.equal(await p.evaluate(()=>backupUntilDate),'2026-10-11');assert.equal(await p.locator('#backup-until-date').inputValue(),'2026-10-12');assert.equal(await p.evaluate(()=>backupMatchCount),57);
});

test('logout and expiry clear dates; late success and 401 cannot change a fresh session date range',async t=>{
 for(const status of [200,401]){const p=await fixture(t,dates);await read(p);await apply(p,'2026-10-11','2026-10-11');await p.evaluate(()=>mode='pending');await p.locator('#reload-backups').click();await p.waitForFunction(()=>pending.length===1);await p.evaluate(()=>{authenticated=false;csrf='';showAuth()});assert.equal(await p.locator('#backup-from-date').inputValue(),'');assert.equal(await p.locator('#apply-backup-date').isDisabled(),true);assert.equal(await p.evaluate(()=>backupDateStart),null);
 await p.evaluate(()=>{authenticated=true;csrf='fresh-session';mode='success';document.getElementById('auth').hidden=true;document.getElementById('workspace').hidden=false;showBackups(true)});await idle(p);await apply(p,'2026-10-12','');const fresh=await ids(p);
 await p.evaluate(({status,old})=>pending[0]({ok:status===200,status,json:async()=>status===200?old:{error:'late expiry'}}),{status,old:dates});await p.waitForTimeout(35);assert.equal(await p.evaluate(()=>authenticated),true);assert.equal(await p.evaluate(()=>backupFromDate),'2026-10-12');assert.deepEqual(await ids(p),fresh);
 await p.evaluate(()=>{fetch=async()=>({ok:false,status:401,json:async()=>({error:'expiry'})})});await p.locator('#reload-backups').click();await p.waitForFunction(()=>!authenticated);assert.equal(await p.locator('#backup-from-date').inputValue(),'');assert.equal(await p.evaluate(()=>backupDateEnd),null);}
});

test('day ranges follow 23-hour and 25-hour daylight-saving days, include last millisecond and exclude next midnight',async t=>{
 for(const [date,start,end,hours] of [['2026-03-08','2026-03-08T05:00:00Z','2026-03-09T04:00:00Z',23],['2026-11-01','2026-11-01T04:00:00Z','2026-11-02T05:00:00Z',25]]){
 const v={...catalog,backups:[dated(0,new Date(Date.parse(start)-1).toISOString()),dated(1,start),dated(2,new Date(Date.parse(end)-1).toISOString()),dated(3,end)]},p=await fixture(t,v,'America/New_York');await read(p);await apply(p,date,date);assert.deepEqual(await ids(p),[entry(1).id,entry(2).id]);assert.equal(await p.evaluate(()=>(backupDateEnd-backupDateStart)/3600000),hours);assert.equal(await p.evaluate(()=>backupDateTimeZone),'America/New_York');}
});

