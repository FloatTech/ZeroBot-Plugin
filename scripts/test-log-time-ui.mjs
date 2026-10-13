// Offline Chrome: synthetic timestamps, mocked APIs, no user browser or live requests.
import assert from 'node:assert/strict';
import fs from 'node:fs/promises';
import path from 'node:path';
import {createRequire} from 'node:module';
import test from 'node:test';
const require=createRequire(import.meta.url);
const modules=process.env.LOCALADMIN_NODE_MODULES || path.join(process.env.USERPROFILE,'.cache/codex-runtimes/codex-primary-runtime/dependencies/node/node_modules');
const {chromium}=require(path.join(modules,'playwright'));
const html=(await fs.readFile('internal/localadmin/assets/index.html','utf8')).replace(/<script\b[^>]*>[\s\S]*?<\/script>/g,'').replace(/<link\b[^>]*>/g,'');
const app=await fs.readFile('internal/localadmin/assets/app.js','utf8'),logs=await fs.readFile('internal/localadmin/assets/logs.js','utf8');
async function fixture(t,timezoneId='Asia/Shanghai') {
 const browser=await chromium.launch({channel:'chrome',headless:true});t.after(()=>browser.close());
 const context=await browser.newContext({timezoneId});await context.route('**/*',r=>r.abort());const page=await context.newPage();
 await page.setContent(html);await page.addScriptTag({content:app});await page.addScriptTag({content:logs});
 await page.evaluate(()=>{
  authenticated=true;csrf='time-fixture';logActive=true;logPaused=true;
  document.getElementById('auth').hidden=true;document.getElementById('workspace').hidden=false;
  for(const s of document.querySelectorAll('.content>section'))s.hidden=s.id!=='logs-view';
  window.requests=[];window.downloads=[];
  window.fetch=async url=>{requests.push(url);return {ok:true,status:200,json:async()=>({entries:[],cursor:'10',instance:'0123456789abcdef',retained:10,capacity:1000}),headers:new Headers({'Content-Type':'text/plain','Content-Length':'7','Content-Disposition':'attachment; filename="zerobot-logs-20261009T000000Z.txt"'}),blob:async()=>new Blob(['fixture'])};};
  URL.createObjectURL=()=> 'blob:fixture';URL.revokeObjectURL=()=>{};HTMLAnchorElement.prototype.click=function(){downloads.push(this.download);};
 });return page;
}
const query=url=>new URL(url,'http://fixture.invalid').searchParams;
async function apply(page,from,until) {await page.locator('#log-from').fill(from);await page.locator('#log-until').fill(until);await page.locator('#log-filter button[type=submit]').click();}

test('local seconds convert to UTC; drafts/export/issues/reload/pause use applied range and clear removes both bounds',async t=>{
 const page=await fixture(t);await page.selectOption('#log-level','warning');await page.locator('#log-keyword').fill('needle');
 await apply(page,'2026-10-09T16:00:01','2026-10-09T16:00:02');
 assert.equal(await page.evaluate(()=>requests.length),0);assert.match(await page.locator('#log-filter-status').textContent(),/16:00:01.*16:00:02/);
 await page.locator('#pause-logs').click();await page.waitForFunction(()=>requests.length===1 && !logRequest);
 let params=query((await page.evaluate(()=>requests))[0]);assert.equal(params.get('from'),'2026-10-09T08:00:01.000Z');assert.equal(params.get('until'),'2026-10-09T08:00:02.000Z');assert.equal(params.get('q'),'needle');assert.equal(params.get('level'),'warning');
 await page.locator('#pause-logs').click();await page.locator('#log-from').fill('2026-10-09T15:00');
 assert.match(await page.locator('#log-filter-status').textContent(),/尚未应用/);
 await page.locator('#export-logs').click();await page.waitForFunction(()=>downloads.length===1);
 params=query((await page.evaluate(()=>requests)).at(-1));assert.equal(params.get('from'),'2026-10-09T08:00:01.000Z');assert.equal(params.get('until'),'2026-10-09T08:00:02.000Z');assert.equal(params.has('after'),false);
 await page.locator('#only-issue-logs').click();assert.equal(await page.locator('#log-from').inputValue(),'2026-10-09T16:00:01');assert.equal(await page.evaluate(()=>logPaused),true);
 await page.locator('#reload-logs').click();assert.equal(await page.evaluate(()=>logUntil),'2026-10-09T16:00:02');
 await page.locator('#clear-log-filter').click();assert.equal(await page.locator('#log-from').inputValue(),'');assert.equal(await page.locator('#log-until').inputValue(),'');assert.equal(await page.evaluate(()=>logPaused),true);
 await page.locator('#pause-logs').click();await page.waitForFunction(()=>requests.length===3 && !logRequest);params=query((await page.evaluate(()=>requests)).at(-1));assert.equal(params.has('from'),false);assert.equal(params.has('until'),false);assert.equal(params.get('level'),'');
});

test('one-sided/equal range works, reversed/impossible/DST times preserve applied conditions and visible rows',async t=>{
 const page=await fixture(t,'America/New_York');await apply(page,'2026-10-09T12:00','');
 assert.equal(await page.evaluate(()=>appliedLogQuery().get('from')),'2026-10-09T16:00:00.000Z');assert.equal(await page.evaluate(()=>appliedLogQuery().has('until')),false);
 await page.evaluate(()=>renderLogs({entries:[{id:'1',time:'2026-10-09T16:00:00Z',level:'info',message:'baseline'}],cursor:'1',instance:'0123456789abcdef',retained:1,capacity:1000}));
 await apply(page,'2026-10-09T13:00','2026-10-09T12:00');assert.equal(await page.locator('#log-filter-error').isVisible(),true);assert.match(await page.locator('#log-filter-error').textContent(),/不能晚于/);
 assert.equal(await page.evaluate(()=>logFrom),'2026-10-09T12:00');assert.equal(await page.locator('#log-records article').count(),1);assert.equal(await page.evaluate(()=>requests.length),0);
 for(const bad of ['2026-02-30T12:00:00','2026-03-08T02:30:00','nonsense']) {await page.evaluate(bad=>applyLogFilters('error','bad',bad,''),bad);assert.equal(await page.locator('#log-filter-error').isVisible(),true);assert.equal(await page.evaluate(()=>logLevel),'');assert.equal(await page.evaluate(()=>logKeyword),'');}
 await apply(page,'','2026-10-09T12:00');assert.equal(await page.evaluate(()=>appliedLogQuery().has('from')),false);assert.equal(await page.evaluate(()=>appliedLogQuery().get('until')),'2026-10-09T16:00:00.000Z');
 await apply(page,'2026-10-09T12:00','2026-10-09T12:00');assert.equal(await page.locator('#log-filter-error').isVisible(),false);
});

test('new time range ignores late rows/downloads; navigation retains drafts; logout resets dates and old 401 cannot reset new login',async t=>{
 const page=await fixture(t);await apply(page,'2026-10-09T16:00','');
 await page.evaluate(()=>{window.pending=[];window.fetch=(url,options)=>{requests.push(url);return new Promise(resolve=>pending.push({resolve,signal:options.signal}));};});
 await page.locator('#pause-logs').click();await page.waitForFunction(()=>pending.length===1);
 await apply(page,'2026-10-09T17:00','');await page.waitForFunction(()=>pending.length===2);assert.equal(await page.evaluate(()=>pending[0].signal.aborted),true);
 await page.evaluate(()=>pending[0].resolve({ok:true,status:200,json:async()=>({entries:[{id:'1',time:'2026-10-09T08:00:00Z',level:'info',message:'OLD RANGE'}],cursor:'1',instance:'0123456789abcdef',retained:1,capacity:1000})}));
 await page.evaluate(()=>pending[1].resolve({ok:true,status:200,json:async()=>({entries:[],cursor:'2',instance:'0123456789abcdef',retained:2,capacity:1000})}));await page.waitForFunction(()=>!logRequest);assert.doesNotMatch(await page.locator('#log-records').textContent(),/OLD RANGE/);
 await page.locator('#pause-logs').click();await page.locator('#export-logs').click();await page.waitForFunction(()=>pending.length===3);
 await page.evaluate(()=>pending[2].resolve({ok:true,status:200,headers:new Headers({'Content-Type':'text/plain','Content-Length':'7','Content-Disposition':'attachment; filename="zerobot-logs-20261009T000000Z.txt"'}),blob:()=>new Promise(resolve=>{window.finishBlob=resolve;})}));await page.waitForFunction(()=>window.finishBlob);
 await apply(page,'','2026-10-09T17:00');await page.evaluate(()=>finishBlob(new Blob(['old'])));assert.equal(await page.evaluate(()=>downloads.length),0);
 await page.locator('#log-from').fill('2026-10-09T15:00');await page.evaluate(()=>{showLogs(false);showLogs(true);});assert.equal(await page.locator('#log-from').inputValue(),'2026-10-09T15:00');assert.equal(await page.evaluate(()=>logFrom),'');
 await page.locator('#pause-logs').click();await page.waitForFunction(()=>pending.length===4);
 await page.evaluate(()=>{authenticated=false;resetLogs();authenticated=true;csrf='new-time-fixture';showLogs(true);});await page.waitForFunction(()=>pending.length===5);
 await page.evaluate(()=>pending[3].resolve({ok:false,status:401,json:async()=>({error:'old session'})}));await page.evaluate(()=>Promise.resolve());
 assert.equal(await page.evaluate(()=>authenticated),true);assert.equal(await page.evaluate(()=>csrf),'new-time-fixture');assert.equal(await page.locator('#log-from').inputValue(),'');assert.equal(await page.locator('#log-until').inputValue(),'');assert.equal(await page.evaluate(()=>logUntil),'');
 await page.evaluate(()=>pending[4].resolve({ok:false,status:401,json:async()=>({error:'expired'})}));await page.waitForFunction(()=>!authenticated);assert.equal(await page.locator('#log-until').inputValue(),'');
});
