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

const ids=p=>p.locator('#backup-list .backup-card small').allTextContents();
const data=(n=200)=>({...catalog,backups:Array.from({length:n},(_,i)=>({...entry(i),createdAt:new Date(Date.UTC(2026,9,11,0,0,n-i)).toISOString(),size:1000+(i%17)}))});

test('all 200 matching records sort before paging, ties are stable and raw byte order survives equal displayed KiB',async t=>{
 const v=data(),p=await fixture(t,v);await read(p);await p.locator('#backup-next').click();await p.selectOption('#backup-sort','oldest');assert.equal(await p.evaluate(()=>backupPage),1);
 const collected=[];for(let i=0;i<10;i++){collected.push(...await ids(p));if(i<9)await p.locator('#backup-next').click()};assert.deepEqual(collected,v.backups.map(e=>e.id).reverse());assert.equal(new Set(collected).size,200);assert.deepEqual(await p.evaluate(()=>backupCatalog.backups.map(e=>e.id)),v.backups.map(e=>e.id));
 for(const order of ['largest','smallest']){await p.selectOption('#backup-sort',order);const expected=v.backups.toSorted((a,b)=>order==='largest'?b.size-a.size:a.size-b.size);assert.deepEqual(await ids(p),expected.slice(0,20).map(e=>e.id));assert.equal(await p.evaluate(()=>backupPage),1)}
 await p.selectOption('#backup-scope-filter','startup-config');assert.equal((await ids(p)).length,10);await p.selectOption('#backup-state-filter','legacy');await p.locator('#backup-search').fill(v.backups[50].id);assert.deepEqual(await ids(p),[v.backups[50].id]);
 await p.locator('#clear-backup-filter').click();assert.equal(await p.locator('#backup-sort').inputValue(),'smallest');await p.selectOption('#backup-sort','default');assert.deepEqual(await ids(p),v.backups.slice(0,20).map(e=>e.id));assert.equal(await p.evaluate(()=>requests.length),1);
});

test('invalid records and unknown ordering fields stay last, zero size is valid and equal keys retain loaded order',async t=>{
 const v={...catalog,backups:[{...entry(0),valid:false,size:999999},{...entry(1),createdAt:'bad-date',size:null},{...entry(2),createdAt:'2026-10-10T23:00:00Z',size:0},{...entry(3),createdAt:'2026-10-10T23:00:00Z',size:0},{...entry(4),size:100},{...entry(5),size:-1},{...entry(6),size:1.5}]},p=await fixture(t,v);await read(p);
 await p.selectOption('#backup-sort','oldest');assert.deepEqual(await ids(p),[2,3,4,5,6,0,1].map(i=>entry(i).id));
 await p.selectOption('#backup-sort','largest');assert.deepEqual(await ids(p),[4,2,3,0,1,5,6].map(i=>entry(i).id));
 await p.selectOption('#backup-sort','smallest');assert.deepEqual(await ids(p),[2,3,4,0,1,5,6].map(i=>entry(i).id));
 assert.equal(await p.locator('.backup-card').filter({hasText:entry(0).id}).locator('button').isDisabled(),true);
 await p.locator('#backup-search').fill('unmatched');assert.equal((await ids(p)).length,0);assert.equal(await p.locator('#backup-next').isDisabled(),true);assert.equal(await p.locator('#backup-prev').isDisabled(),true);assert.equal(await p.evaluate(()=>requests.length),1);
});

test('download uses the selected all-page order and fixed sort field; sorting preserves restore/cleanup tickets and checks',async t=>{
 const v=data(65),p=await fixture(t,v);await read(p);await p.evaluate(()=>{backupPreview={backup:data.backups[3],token:'fixture-private-ticket',changed:true,expiresAt:Math.floor(Date.now()/1000)+600};backupCleanupPreview={token:'fixture-private-cleanup',candidates:[{id:'fixture'}],expiresAt:Math.floor(Date.now()/1000)+600};document.getElementById('restore-backup-id').textContent=data.backups[3].id;document.getElementById('restore-confirm').checked=true;document.getElementById('cleanup-confirm').checked=true;renderBackupExpiry()});
 await p.selectOption('#backup-sort','oldest');await p.selectOption('#backup-scope-filter','control-db');await p.locator('#backup-next').click();const downloaded=await download(p);
 assert.equal(downloaded.value.filters.sort,'oldest');assert.deepEqual(downloaded.value.backups.map(e=>e.id),v.backups.slice(0,45).map(e=>e.id).reverse());assert.doesNotMatch(downloaded.text,/fixture-private|token|config|SHA256/);assert.equal(await p.evaluate(()=>backupPage),2);assert.equal(await p.locator('#restore-confirm').isChecked(),true);assert.equal(await p.locator('#cleanup-confirm').isChecked(),true);assert.equal(await p.locator('#restore-backup-id').textContent(),v.backups[3].id);assert.equal(await p.evaluate(()=>backupPreview.token),'fixture-private-ticket');assert.equal(await p.evaluate(()=>backupCleanupPreview.token),'fixture-private-cleanup');assert.equal(await p.evaluate(()=>requests.length),1);
});

test('refresh failures and navigation preserve sort, shrinking refresh corrects pages, busy and invalid changes are rejected',async t=>{
 const p=await fixture(t,data(65));assert.equal(await p.locator('#backup-sort').isDisabled(),true);await p.evaluate(()=>{backupActive=true;const s=document.getElementById('backup-sort');s.value='largest';s.dispatchEvent(new Event('change'))});assert.equal(await p.locator('#backup-sort').inputValue(),'default');await read(p);await p.selectOption('#backup-sort','oldest');await p.locator('#backup-next').click();
 const previous=await ids(p);await p.evaluate(()=>mode='failure');await p.locator('#reload-backups').click();await idle(p);assert.deepEqual(await ids(p),previous);assert.equal(await p.locator('#backup-sort').inputValue(),'oldest');assert.equal(await p.evaluate(()=>backupPage),2);assert.match(await p.locator('#backup-filter-count').textContent(),/刷新失败/);const stale=await download(p);assert.equal(stale.value.source.lastReadFailed,true);assert.equal(stale.value.filters.sort,'oldest');
 await p.evaluate(()=>{showBackups(false);mode='success';data={...data,backups:data.backups.slice(0,2)};showBackups(true)});await idle(p);assert.equal(await p.evaluate(()=>backupPage),1);assert.equal(await p.locator('#backup-sort').inputValue(),'oldest');
 await p.evaluate(()=>mode='pending');await p.locator('#reload-backups').click();await p.waitForFunction(()=>pending.length===1);assert.equal(await p.locator('#backup-sort').isDisabled(),true);await p.evaluate(()=>{const s=document.getElementById('backup-sort');s.value='largest';s.dispatchEvent(new Event('change'))});assert.equal(await p.locator('#backup-sort').inputValue(),'oldest');
 await p.evaluate(()=>pending[0]({ok:true,status:200,json:async()=>data}));await idle(p);await p.evaluate(()=>{const s=document.getElementById('backup-sort');s.value='unknown';s.dispatchEvent(new Event('change'))});assert.equal(await p.locator('#backup-sort').inputValue(),'oldest');
});

test('expired sessions reset sorting and old success/401 cannot alter a new session sorted catalog',async t=>{
 for(const status of [200,401]){const p=await fixture(t,data(65));await read(p);await p.selectOption('#backup-sort','oldest');await p.evaluate(()=>mode='pending');await p.locator('#reload-backups').click();await p.waitForFunction(()=>pending.length===1);
 await p.evaluate(()=>{authenticated=false;csrf='';showAuth()});assert.equal(await p.locator('#backup-sort').inputValue(),'default');assert.equal(await p.locator('#backup-sort').isDisabled(),true);
 await p.evaluate(()=>{authenticated=true;csrf='fresh-session';mode='success';data={...data,backups:data.backups.slice(0,3)};document.getElementById('auth').hidden=true;document.getElementById('workspace').hidden=false;showBackups(true)});await idle(p);await p.selectOption('#backup-sort','largest');const fresh=await ids(p);
 await p.evaluate(({status,old})=>pending[0]({ok:status===200,status,json:async()=>status===200?old:{error:'late expiry'}}),{status,old:data()});await p.waitForTimeout(40);assert.equal(await p.evaluate(()=>authenticated),true);assert.equal(await p.locator('#backup-sort').inputValue(),'largest');assert.deepEqual(await ids(p),fresh);
 await p.evaluate(()=>{mode='expired';fetch=async()=>({ok:false,status:401,json:async()=>({error:'current expiry'})})});await p.locator('#reload-backups').click();await p.waitForFunction(()=>!authenticated);assert.equal(await p.locator('#backup-sort').inputValue(),'default');assert.equal(await p.evaluate(()=>backupCatalog),null);}
});

