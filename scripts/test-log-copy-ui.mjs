// Offline Chrome: synthetic redacted log pages, mocked API and no live robot.
import assert from 'node:assert/strict';import fs from 'node:fs/promises';import path from 'node:path';import {createRequire} from 'node:module';import test,{after} from 'node:test';
const require=createRequire(import.meta.url),modules=process.env.LOCALADMIN_NODE_MODULES||path.join(process.env.USERPROFILE,'.cache/codex-runtimes/codex-primary-runtime/dependencies/node/node_modules');const {chromium}=require(path.join(modules,'playwright'));const browser=await chromium.launch({channel:'chrome',headless:true});after(()=>browser.close());
const html=(await fs.readFile('internal/localadmin/assets/index.html','utf8')).replace(/<script\b[^>]*>[\s\S]*?<\/script>/g,'').replace(/<link\b[^>]*>/g,'');const app=await fs.readFile('internal/localadmin/assets/app.js','utf8'),logs=await fs.readFile('internal/localadmin/assets/logs.js','utf8');
const levels=(overrides={})=>({panic:0,fatal:0,error:0,warning:0,info:0,debug:0,trace:0,other:0,...overrides});
const row=id=>({id:String(id),time:'2026-10-10T12:00:00Z',level:'info',message:'redacted fixture [已隐藏]'});
const page=(overrides={})=>({entries:[row(499),row(500)],cursor:'500',instance:'0123456789abcdef',retained:500,capacity:1000,summary:{matched:350,levels:levels({info:345,warning:4,error:1})},...overrides});
async function fixture(t){const context=await browser.newContext({timezoneId:'Asia/Shanghai'});t.after(()=>context.close());await context.route('**/*',r=>r.abort());const p=await context.newPage();await p.setContent(html);await p.addScriptTag({content:app});await p.addScriptTag({content:logs});await p.evaluate(()=>{authenticated=true;csrf='summary-session';logActive=true;logPaused=true;document.getElementById('pause-logs').textContent='恢复更新';document.getElementById('auth').hidden=true;document.getElementById('workspace').hidden=false;for(const s of document.querySelectorAll('.content>section'))s.hidden=s.id!=='logs-view';window.showOverview=()=>{};window.requests=[];window.replies=[];window.pending=[];window.fetch=async(url,options)=>{requests.push({url,method:options.method||'GET'});const reply=replies.shift();if(!reply)return new Promise(resolve=>pending.push(resolve));return {ok:!reply.error,status:reply.status||200,json:async()=>reply};};});return p;}
const state=p=>p.locator('#log-summary-state').textContent();
const badges=p=>p.locator('#log-summary-levels').textContent();
const idle=p=>p.waitForFunction(()=>!logRequest);

async function ready(t,value=page()){const p=await fixture(t);await p.evaluate(value=>{window.written=[];window.copyPending=[];window.copyMode='success';Object.defineProperty(navigator,'clipboard',{configurable:true,value:{writeText:async text=>{written.push(text);if(copyMode==='pending')return new Promise((resolve,reject)=>copyPending.push({resolve,reject}));if(copyMode==='failure')throw Error('fixture clipboard rejection')}}});renderLogs(value)},value);return p;}
const button=(p,id)=>p.locator(`.log-entry[data-id="${id}"] .copy-log-entry`);
const copyStatus=(p,id)=>p.locator(`.log-entry[data-id="${id}"] .log-copy-actions [role=status]`);
const copied=async(p,id)=>p.waitForFunction(id=>document.querySelector(`.log-entry[data-id="${id}"] .log-copy-actions [role=status]`).textContent==='已复制本条日志。',id);

test('copy uses displayed local time, level and redacted Unicode/multiline text only, without changing log state or API calls',async t=>{
 const entry={...row(1),level:'error',message:'中文鲸鱼🐳\n[已隐藏] token=[已隐藏] <img src=x> [截断]',rawMessage:'fixture-private-raw',token:'fixture-private-token'},p=await ready(t,page({entries:[entry],cursor:'1'}));await p.locator('#log-autoscroll').uncheck();
 const before=await p.evaluate(()=>({cursor:logCursor,instance:logInstance,paused:logPaused,level:logLevel,keyword:logKeyword,summary:document.getElementById('log-summary-state').textContent,seen:logSeen.size}));await button(p,1).click();await copied(p,1);const text=(await p.evaluate(()=>written))[0],display=await p.locator('.log-meta time').textContent();assert.equal(text,display+' [ERROR]\n'+entry.message);assert.doesNotMatch(text,/fixture-private/);assert.equal(await p.locator('#log-records img').count(),0);assert.equal(await button(p,1).isEnabled(),true);
 assert.deepEqual(await p.evaluate(()=>({cursor:logCursor,instance:logInstance,paused:logPaused,level:logLevel,keyword:logKeyword,summary:document.getElementById('log-summary-state').textContent,seen:logSeen.size})),before);assert.equal(await p.locator('#log-autoscroll').isChecked(),false);assert.equal(await p.evaluate(()=>requests.length),0);
 await p.evaluate(()=>{logPaused=false;document.getElementById('pause-logs').textContent='暂停更新'});await button(p,1).click();await copied(p,1);assert.equal(await p.evaluate(()=>written.length),2);assert.equal(await p.evaluate(()=>logPaused),false);assert.equal(await p.evaluate(()=>logTimer),null);
});

test('one entry blocks duplicate writes while other entries have independent copy status; tail append preserves pending copies',async t=>{
 const p=await ready(t);await p.evaluate(()=>copyMode='pending');await button(p,499).click();await p.waitForFunction(()=>copyPending.length===1);assert.equal(await button(p,499).isDisabled(),true);await p.evaluate(()=>copyLogEntry(document.querySelector('.copy-log-entry')));assert.equal(await p.evaluate(()=>written.length),1);await button(p,500).click();await p.waitForFunction(()=>copyPending.length===2);
 await p.evaluate(value=>renderLogs(value),page({entries:[row(501)],cursor:'501'}));assert.equal(await button(p,499).isDisabled(),true);assert.equal(await button(p,501).isEnabled(),true);
 await p.evaluate(()=>copyPending[1].resolve());await copied(p,500);assert.equal(await copyStatus(p,499).textContent(),'正在复制…');await p.evaluate(()=>copyPending[0].resolve());await copied(p,499);assert.equal(await button(p,499).isEnabled(),true);assert.equal(await p.evaluate(()=>requests.length),0);
});

test('rejected or unavailable clipboard selects only the displayed body, supports retry and clears owned selection on lifecycle changes',async t=>{
 const entry={...row(1),message:'手动复制🐳\n第二行 [已隐藏]'},p=await ready(t,page({entries:[entry]}));await p.evaluate(()=>copyMode='failure');await button(p,1).click();await p.waitForFunction(()=>window.getSelection().toString().includes('手动复制'));assert.equal(await p.evaluate(()=>window.getSelection().toString()),entry.message);assert.match(await copyStatus(p,1).textContent(),/Ctrl\+C/);assert.equal(await button(p,1).isEnabled(),true);await p.evaluate(()=>{stopLogRequests();copyMode='success'});assert.equal(await p.evaluate(()=>window.getSelection().toString()),'');await button(p,1).click();await copied(p,1);
 await p.evaluate(()=>Object.defineProperty(navigator,'clipboard',{configurable:true,value:undefined}));await button(p,1).click();await p.waitForFunction(()=>window.getSelection().toString().includes('手动复制'));assert.equal(await p.evaluate(()=>window.getSelection().toString()),entry.message);
 await p.locator('[data-view=overview]').click();assert.equal(await p.evaluate(()=>window.getSelection().toString()),'');assert.equal(await button(p,1).isDisabled(),true);assert.equal(await copyStatus(p,1).textContent(),'');
});

test('filters, reload, navigation, visibility and pause invalidate late success/failure without choosing stale text',async t=>{
 for(const action of ['filter','reload','navigation','hidden','pause']){for(const failure of [false,true]){const p=await ready(t);await p.evaluate(()=>copyMode='pending');await button(p,499).click();await p.waitForFunction(()=>copyPending.length===1);
 if(action==='filter')await p.evaluate(()=>applyLogFilters('error','new'));
 if(action==='reload')await p.locator('#reload-logs').click();
 if(action==='navigation')await p.locator('[data-view=overview]').click();
 if(action==='hidden')await p.evaluate(()=>{Object.defineProperty(document,'hidden',{configurable:true,value:true});document.dispatchEvent(new Event('visibilitychange'))});
 if(action==='pause')await p.evaluate(()=>{logPaused=false;document.getElementById('pause-logs').click()});
 await p.evaluate(failure=>failure?copyPending[0].reject(Error('late clipboard failure')):copyPending[0].resolve(),failure);await p.waitForTimeout(30);assert.equal(await p.evaluate(()=>window.getSelection().toString()),'');assert.doesNotMatch(await p.locator('#log-records').textContent(),/已复制|Ctrl\+C|正在复制/);
 assert.equal(await p.evaluate(()=>requests.length),0);}}
});

test('capacity eviction and instance reset detach old entries; late failure cannot select or overwrite replacement results',async t=>{
 for(const reset of [false,true]){const p=await ready(t,page({entries:[row(1)],cursor:'1'}));await p.evaluate(()=>copyMode='pending');await button(p,1).click();await p.waitForFunction(()=>copyPending.length===1);
 await p.evaluate(({reset,value})=>renderLogs({...value,reset,entries:reset?[{...value.entries[0],id:'1',message:'replacement [已隐藏]'}]:Array.from({length:1000},(_,i)=>({...value.entries[0],id:String(i+2)}))}),{reset,value:page()});assert.equal(await p.locator('.log-entry').count(),reset?1:1000);if(!reset)assert.equal(await button(p,1).count(),0);
 await p.evaluate(()=>copyPending[0].reject(Error('old detached rejection')));await p.waitForTimeout(30);assert.equal(await p.evaluate(()=>window.getSelection().toString()),'');assert.doesNotMatch(await p.locator('#log-records').textContent(),/Ctrl\+C|正在复制/);assert.equal(await p.locator('.copy-log-entry:disabled').count(),0);assert.equal(await p.evaluate(()=>requests.length),0);}
});

test('logout and current expiry clear old copies; late completions cannot alter a new session copy or enable its busy button',async t=>{
 for(const failure of [false,true]){const p=await ready(t);await p.evaluate(()=>copyMode='pending');await button(p,499).click();await p.waitForFunction(()=>copyPending.length===1);await p.evaluate(()=>{authenticated=false;csrf='';showAuth()});assert.equal(await p.locator('.log-entry').count(),0);
 await p.evaluate(value=>{authenticated=true;csrf='new-session';logActive=true;logPaused=true;document.getElementById('auth').hidden=true;document.getElementById('workspace').hidden=false;renderLogs(value)},page());await button(p,499).click();await p.waitForFunction(()=>copyPending.length===2);
 await p.evaluate(failure=>failure?copyPending[0].reject(Error('old clipboard rejection')):copyPending[0].resolve(),failure);await p.waitForTimeout(30);assert.equal(await button(p,499).isDisabled(),true);assert.equal(await copyStatus(p,499).textContent(),'正在复制…');assert.equal(await p.evaluate(()=>window.getSelection().toString()),'');await p.evaluate(()=>copyPending[1].resolve());await copied(p,499);
 await p.evaluate(()=>{replies.push({error:'current expiry',status:401});readLogsOnce()});await p.waitForFunction(()=>!authenticated);assert.equal(await p.locator('.log-entry').count(),0);assert.equal(await p.evaluate(()=>window.getSelection().toString()),'');}
});

