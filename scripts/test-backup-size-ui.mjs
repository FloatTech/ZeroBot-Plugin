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
async function fixture(t,data=catalog){const context=await browser.newContext({acceptDownloads:true});t.after(()=>context.close());await context.route('**/*',r=>r.abort());const p=await context.newPage();await p.setContent(html);await p.addScriptTag({content:app});await p.addScriptTag({content:backups});await p.evaluate(data=>{authenticated=true;csrf='catalog-session';window.showOverview=()=>{};window.showLogs=()=>{};document.getElementById('auth').hidden=true;document.getElementById('workspace').hidden=false;document.getElementById('overview-view').hidden=true;document.getElementById('backups-view').hidden=false;window.data=data;window.requests=[];window.pending=[];window.mode='success';window.revoked=[];const revoke=URL.revokeObjectURL.bind(URL);URL.revokeObjectURL=u=>{revoked.push(u);revoke(u)};window.fetch=async(url,options)=>{requests.push({url,method:options.method||'GET'});if(mode==='pending')return new Promise(resolve=>pending.push(resolve));return {ok:mode==='success',status:mode==='success'?200:503,json:async()=>mode==='success'?window.data:{error:'fixture read failed'}}};},data);return p;}
const idle=p=>p.waitForFunction(()=>!backupBusy);
const read=async p=>{await p.evaluate(()=>showBackups(true));await idle(p);};
async function download(p){const [d]=await Promise.all([p.waitForEvent('download'),p.locator('#download-backup-catalog').click()]);const text=await fs.readFile(await d.path(),'utf8');return {d,text,value:JSON.parse(text)};}

const loaded=p=>p.locator('#backup-size-loaded').textContent();
const matched=p=>p.locator('#backup-size-matched').textContent();
const scopes=p=>p.locator('#backup-size-scopes').textContent();

test('summary covers all 200 records and all filtered pages; scopes, sorting and pagination cannot change totals',async t=>{
 const v={...catalog,backups:Array.from({length:200},(_,i)=>({...entry(i),size:1024}))},p=await fixture(t,v);assert.equal(await p.locator('#backup-size-summary').isHidden(),true);await read(p);
 assert.match(await loaded(p),/已加载：200 份.*200\.00 KiB（204800 字节）.*无法核对：0 份/);assert.match(await matched(p),/200 份.*204800 字节/);assert.equal(await p.locator('#backup-size-scopes article').count(),4);assert.match(await scopes(p),/控制数据库.*45 份.*46080 字节/);assert.match(await p.locator('#backup-size-status').textContent(),/最近成功读取.*最新200份.*未加载记录/);
 const original=await loaded(p);await p.selectOption('#backup-scope-filter','control-db');assert.equal(await loaded(p),original);assert.match(await matched(p),/45 份.*45\.00 KiB（46080 字节）/);const filtered=await matched(p);
 await p.selectOption('#backup-page-size','10');await p.locator('#backup-next').click();await p.selectOption('#backup-sort','largest');assert.equal(await loaded(p),original);assert.equal(await matched(p),filtered);
 await p.selectOption('#backup-scope-filter','startup-config');assert.match(await matched(p),/10 份.*10240 字节/);await p.locator('#backup-search').fill('no-match');assert.match(await matched(p),/0 份.*0 字节.*无法核对：0 份/);await p.locator('#clear-backup-filter').click();assert.match(await matched(p),/200 份.*204800 字节/);assert.equal(await p.evaluate(()=>requests.length),1);
});

test('zero size is known, invalid or unreadable sizes are excluded, huge integer totals remain exact and metadata never enters summary',async t=>{
 const v={...catalog,more:false,backups:[{...entry(0),size:0},{...entry(1),size:1024},{...entry(2),size:-1},{...entry(3),size:1.5},{...entry(4),size:'999999'},{...entry(5),valid:false,size:999999},{...entry(6),size:null},{...entry(7),scope:'<private-scope>',valid:false,size:0}]},p=await fixture(t,v);await read(p);
 assert.match(await loaded(p),/8 份.*1\.00 KiB（1024 字节）.*无法核对：6 份/);assert.match(await scopes(p),/异常\/未知范围.*1 份.*无法核对：1 份/);assert.doesNotMatch(await p.locator('#backup-size-summary').textContent(),/private|999999|SHA256|config|token|<img/);assert.equal(await p.locator('#backup-size-summary img').count(),0);
 await p.evaluate(()=>{mode='success';data={backups:[{...data.backups[0],size:Number.MAX_SAFE_INTEGER},{...data.backups[1],size:Number.MAX_SAFE_INTEGER}],scopes:{},more:false}});await p.locator('#reload-backups').click();await idle(p);assert.match(await loaded(p),/18014398509481982 字节/);assert.match(await loaded(p),/16\.00 PiB/);assert.doesNotMatch(await loaded(p),/18014398509481984/);
 await p.evaluate(()=>{data={backups:[],scopes:{},more:false}});await p.locator('#reload-backups').click();await idle(p);assert.match(await loaded(p),/0 份.*0 字节.*无法核对：0 份/);
});

test('failed refresh retains and labels old summary, successful refresh replaces it, reading does not disturb tickets or confirmations',async t=>{
 const p=await fixture(t);await read(p);const original=await loaded(p);await p.evaluate(()=>mode='failure');await p.locator('#reload-backups').click();await idle(p);assert.equal(await loaded(p),original);assert.match(await p.locator('#backup-size-status').textContent(),/上次读取结果，刷新失败/);
 await p.evaluate(()=>{mode='success';data={...data,backups:data.backups.slice(0,1),more:false}});await p.locator('#reload-backups').click();await idle(p);assert.match(await loaded(p),/1 份.*512 字节/);assert.doesNotMatch(await p.locator('#backup-size-status').textContent(),/刷新失败|未加载记录/);
 await p.evaluate(()=>{backupPreview={token:'fixture-private-restore',changed:true,expiresAt:Math.floor(Date.now()/1000)+600};backupCleanupPreview={token:'fixture-private-cleanup',candidates:[{id:'fixture'}],expiresAt:Math.floor(Date.now()/1000)+600};document.getElementById('restore-confirm').checked=true;document.getElementById('cleanup-confirm').checked=true;renderBackupExpiry()});
 const count=await p.evaluate(()=>requests.length);await p.selectOption('#backup-state-filter','legacy');assert.equal(await p.locator('#restore-confirm').isChecked(),true);assert.equal(await p.locator('#cleanup-confirm').isChecked(),true);assert.deepEqual(await p.evaluate(()=>[backupPreview.token,backupCleanupPreview.token]),['fixture-private-restore','fixture-private-cleanup']);assert.equal(await p.evaluate(()=>requests.length),count);
 await p.evaluate(()=>{showBackups(false);showBackups(true)});await idle(p);assert.match(await loaded(p),/512 字节/);
});

test('initial failure remains unknown and logout/expiry clears totals; stale replies cannot replace a new session summary',async t=>{
 for(const status of [200,401]){
  const p=await fixture(t);await p.evaluate(()=>mode='failure');await read(p);assert.equal(await p.locator('#backup-size-summary').isHidden(),true);assert.equal(await loaded(p),'');await p.evaluate(()=>mode='success');await p.locator('#reload-backups').click();await idle(p);await p.evaluate(()=>mode='pending');await p.locator('#reload-backups').click();await p.waitForFunction(()=>pending.length===1);
  await p.evaluate(()=>{authenticated=false;csrf='';showAuth()});assert.equal(await p.locator('#backup-size-summary').isHidden(),true);assert.equal(await loaded(p),'');assert.equal(await matched(p),'');assert.equal(await scopes(p),'');
  await p.evaluate(()=>{authenticated=true;csrf='fresh-session';mode='success';data={...data,backups:[{...data.backups[0],size:42}],more:false};document.getElementById('auth').hidden=true;document.getElementById('workspace').hidden=false;showBackups(true)});await idle(p);assert.match(await loaded(p),/42 字节/);
  await p.evaluate(({status,old})=>pending[0]({ok:status===200,status,json:async()=>status===200?old:{error:'old expiry'}}),{status,old:catalog});await p.waitForTimeout(40);assert.equal(await p.evaluate(()=>authenticated),true);assert.match(await loaded(p),/42 字节/);assert.doesNotMatch(await loaded(p),/65 份/);
  await p.evaluate(()=>{fetch=async()=>({ok:false,status:401,json:async()=>({error:'current expiry'})})});await p.locator('#reload-backups').click();await p.waitForFunction(()=>!authenticated);assert.equal(await p.locator('#backup-size-summary').isHidden(),true);assert.equal(await loaded(p),'');
 }
});

