// Offline Chrome: synthetic clock/logs, mocked APIs, no live site or user session.
import assert from 'node:assert/strict';
import fs from 'node:fs/promises';
import path from 'node:path';
import {createRequire} from 'node:module';
import test,{after} from 'node:test';
const require=createRequire(import.meta.url);
const modules=process.env.LOCALADMIN_NODE_MODULES || path.join(process.env.USERPROFILE,'.cache/codex-runtimes/codex-primary-runtime/dependencies/node/node_modules');
const {chromium}=require(path.join(modules,'playwright'));
const browser=await chromium.launch({channel:'chrome',headless:true});after(()=>browser.close());
const html=(await fs.readFile('internal/localadmin/assets/index.html','utf8')).replace(/<script\b[^>]*>[\s\S]*?<\/script>/g,'').replace(/<link\b[^>]*>/g,'');
const app=await fs.readFile('internal/localadmin/assets/app.js','utf8'),logs=await fs.readFile('internal/localadmin/assets/logs.js','utf8');
async function fixture(t,timezoneId='Asia/Shanghai') {
 const context=await browser.newContext({timezoneId});t.after(()=>context.close());await context.route('**/*',r=>r.abort());
 const p=await context.newPage();await p.setContent(html);await p.addScriptTag({content:app});await p.addScriptTag({content:logs});
 await p.evaluate(()=>{
  authenticated=true;csrf='range-session';logActive=true;logPaused=true;window.showOverview=()=>{};
  document.getElementById('auth').hidden=true;document.getElementById('workspace').hidden=false;
  for(const s of document.querySelectorAll('.content>section'))s.hidden=s.id!=='logs-view';
  window.clock=Date.parse('2026-10-11T00:05:30.987+08:00');Date.now=()=>clock;
  window.requests=[];window.downloads=[];
  window.sample={entries:[],cursor:'10',instance:'0123456789abcdef',retained:10,capacity:1000,summary:{matched:3,levels:{panic:0,fatal:0,error:0,warning:3,info:0,debug:0,trace:0,other:0}}};
  window.fetch=async(url,options)=>{requests.push({url,method:options.method||'GET'});return {ok:true,status:200,json:async()=>sample,headers:new Headers({'Content-Type':'text/plain','Content-Length':'7','Content-Disposition':'attachment; filename="zerobot-logs-20261011T000000Z.txt"'}),blob:async()=>new Blob(['fixture'])};};
  URL.createObjectURL=()=> 'blob:fixture';URL.revokeObjectURL=()=>{};HTMLAnchorElement.prototype.click=function(){downloads.push(this.download);};
 });return p;
}
const params=url=>Object.fromEntries(new URL(url,'http://fixture.invalid').searchParams);
const range=p=>p.evaluate(()=>Object.fromEntries(appliedLogQuery()));
const idle=p=>p.waitForFunction(()=>!logRequest);

test('presets cross midnight with exact elapsed duration, preserve applied filters and pause, time-only clear differs from clear-all',async t=>{
 const p=await fixture(t);await p.evaluate(()=>applyLogFilters('warning','saved keyword'));
 for(const [id,minutes] of [['15m',15],['1h',60],['24h',1440]]) {
  await p.selectOption('#log-level','error');await p.locator('#log-keyword').fill('unapplied');await p.locator('#log-from').fill('2026-01-01T00:00');
  await p.locator('#log-recent-'+id).click();const q=await range(p);
  assert.equal(Date.parse(q.until)-Date.parse(q.from),minutes*60000);assert.equal(q.until,'2026-10-10T16:05:30.000Z');assert.equal(q.level,'warning');assert.equal(q.q,'saved keyword');
  assert.equal(await p.locator('#log-until').inputValue(),'2026-10-11T00:05:30');assert.equal(await p.locator('#log-keyword').inputValue(),'saved keyword');assert.equal(await p.evaluate(()=>logPaused),true);
  assert.doesNotMatch(await p.locator('#log-filter-status').textContent(),/尚未应用/);
 }
 assert.equal(await p.evaluate(()=>requests.length),0);await p.locator('#log-any-time').click();assert.deepEqual(await range(p),{level:'warning',q:'saved keyword'});
 await p.locator('#clear-log-filter').click();assert.deepEqual(await range(p),{level:'',q:''});assert.equal(await p.evaluate(()=>requests.length),0);
});

test('poll, summary, reload, navigation and export retain the fixed interval; another click advances it',async t=>{
 const p=await fixture(t);await p.evaluate(()=>applyLogFilters('warning','saved'));await p.locator('#log-recent-1h').click();const original=await range(p);
 await p.locator('#pause-logs').click();await idle(p);assert.match(await p.locator('#log-summary-state').textContent(),/符合已应用筛选：3/);
 await p.locator('#pause-logs').click();await p.evaluate(()=>clock+=600000);await p.locator('#reload-logs').click();await p.locator('[data-view=overview]').click();await p.locator('[data-view=logs]').click();assert.deepEqual(await range(p),original);
 await p.locator('#export-logs').click();await p.waitForFunction(()=>downloads.length===1);const requests=await p.evaluate(()=>requests);
 for(const r of requests) {const q=params(r.url);assert.equal(q.from,original.from);assert.equal(q.until,original.until);assert.equal(q.level,original.level);assert.equal(q.q,original.q);assert.equal(r.method,'GET');}
 assert.equal(params(requests.at(-1).url).after,undefined);await p.locator('#log-recent-1h').click();assert.equal(Date.parse((await range(p)).until)-Date.parse(original.until),600000);
 assert.equal(await p.evaluate(()=>logPaused),true);assert.match(await p.locator('#log-summary-state').textContent(),/读取当前筛选后显示/);
});

test('range switches cancel reads/downloads and late replies cannot refill results or expire a new session',async t=>{
 const p=await fixture(t);await p.evaluate(()=>{window.pending=[];window.fetch=(url,options)=>{requests.push({url,method:options.method||'GET'});return new Promise(resolve=>pending.push({resolve,signal:options.signal}));};});
 await p.locator('#pause-logs').click();await p.locator('#export-logs').click();await p.waitForFunction(()=>pending.length===2);
 await p.locator('#log-recent-15m').click();await p.waitForFunction(()=>pending.length===3);assert.equal(await p.evaluate(()=>pending[0].signal.aborted&&pending[1].signal.aborted),true);
 await p.evaluate(()=>{pending[0].resolve({ok:true,status:200,json:async()=>({...sample,entries:[{id:'1',time:'2026-10-10T16:00:00Z',level:'info',message:'STALE'}]})});pending[1].resolve({ok:true,status:200,headers:new Headers({'Content-Type':'text/plain','Content-Length':'7','Content-Disposition':'attachment; filename="zerobot-logs-20261011T000000Z.txt"'}),blob:async()=>new Blob(['fixture'])});});
 await p.waitForTimeout(30);assert.equal(await p.locator('.log-entry').count(),0);assert.equal(await p.evaluate(()=>downloads.length),0);
 await p.evaluate(()=>{authenticated=false;csrf='';showAuth();authenticated=true;csrf='new-session';logActive=true;logPaused=true;applyLogFilters('warning','new');pending[2].resolve({ok:false,status:401,json:async()=>({error:'expired old session'})});});
 await p.waitForTimeout(30);assert.equal(await p.evaluate(()=>authenticated),true);assert.deepEqual(await range(p),{level:'warning',q:'new'});
 await p.evaluate(()=>{authenticated=false;showAuth();applyRecentLogRange(15)});assert.deepEqual(await range(p),{level:'',q:''});assert.equal(await p.locator('#log-from').inputValue(),'');assert.equal(await p.locator('#log-until').inputValue(),'');
});

test('DST spans use elapsed time; repeated-hour endpoints and invalid clocks preserve original rows and filters',async t=>{
 const p=await fixture(t,'America/New_York');
 for(const timestamp of ['2026-03-08T16:00:00Z','2026-11-01T17:00:00Z']) {
  await p.evaluate(s=>clock=Date.parse(s),timestamp);await p.locator('#log-recent-24h').click();const q=await range(p);assert.equal(Date.parse(q.until)-Date.parse(q.from),86400000);assert.equal(q.until,timestamp.replace('Z','.000Z'));
 }
 const original=await range(p);await p.evaluate(()=>renderLogs({...sample,entries:[{id:'1',time:'2026-11-01T17:00:00Z',level:'warning',message:'preserved'}]}));
 for(const timestamp of ['2026-11-01T06:30:00Z','invalid']) {
  await p.evaluate(s=>clock=Date.parse(s),timestamp);await p.locator('#log-recent-15m').click();assert.equal(await p.locator('#log-filter-error').isVisible(),true);assert.deepEqual(await range(p),original);assert.equal(await p.locator('.log-entry').count(),1);
 }
 assert.equal(await p.evaluate(()=>requests.length),0);await p.evaluate(()=>clock=Date.parse('2026-11-01T17:00:00Z'));await p.locator('#log-recent-15m').click();assert.equal(await p.locator('#log-filter-error').isVisible(),false);
});
