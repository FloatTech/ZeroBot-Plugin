// Offline Chrome with synthetic plugins and mocked API responses only.
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
const app=await fs.readFile('internal/localadmin/assets/app.js','utf8');
const state=(mode='inherit',group='0')=>({effectiveEnabled:mode!=='disabled',defaultEnabled:true,globalMode:group==='0'?mode:'inherit',groupMode:group==='0'?'inherit':mode,source:mode==='inherit'?'default':group==='0'?'global':'group',revision:(mode==='inherit'?'a':'b').repeat(64)});
const fixtures=(n=21)=>Array.from({length:n},(_,i)=>({name:'plugin-'+String(i).padStart(3,'0'),brief:'功能 '+i,usage:'/unique-'+i,defaultEnabledAtStartup:true,state:state()}));
const saved=(i,group='0',mode='disabled')=>({name:'plugin-'+String(i).padStart(3,'0'),groupId:group,state:state(mode,group)});
const card=(p,i)=>p.locator(`[data-plugin="plugin-${String(i).padStart(3,'0')}"]`);
const idle=p=>p.waitForFunction(()=>!pluginSaving&&!pluginScopeBusy);
async function fixture(t,n=21){const context=await browser.newContext();t.after(()=>context.close());await context.route('**/*',r=>r.abort());const p=await context.newPage();await p.setContent(html);await p.addScriptTag({content:app});await p.evaluate(data=>{
 authenticated=true;csrf='batch-session';plugins=data;pluginControlAvailable=true;window.showOverview=()=>{};window.showLogs=()=>{};window.showBackups=()=>{};document.getElementById('auth').hidden=true;document.getElementById('workspace').hidden=false;document.getElementById('overview-view').hidden=true;document.getElementById('plugins-view').hidden=false;window.requests=[];window.replies=[];window.pending=[];
 window.fetch=async(url,options)=>{requests.push({url,body:options.body?JSON.parse(options.body):null});const v=replies.shift();if(!v)throw Error('missing fixture');if(v.throw)throw Error(v.throw);if(v.pending)return new Promise((resolve,reject)=>{pending.push({resolve,reject,signal:options.signal});options.signal?.addEventListener('abort',()=>reject(Object.assign(Error('aborted'),{name:'AbortError'})),{once:true});});return {ok:!v.error,status:v.status||200,json:async()=>v};};renderPluginScope();renderPluginBusy();renderPlugins();
 },fixtures(n));return p;}
async function draft(p,ids){await p.evaluate(ids=>{for(const i of ids){const p=plugins[i];pluginDrafts.set(p.name,{mode:'disabled',revision:p.state.revision});}renderPlugins();},ids);}
async function start(p,accept=true){const event=p.waitForEvent('dialog'),click=p.locator('#save-all-plugin-drafts').click();const dialog=await event;const message=dialog.message();if(accept)await dialog.accept();else await dialog.dismiss();await click;return message;}
async function resolve(p,index,value,status=200){await p.evaluate(({index,value,status})=>pending[index].resolve({ok:status===200,status,json:async()=>value}),{index,value,status});}

test('confirmation includes hidden all-page drafts and loaded scope; cancellation and stale confirmation never write',async t=>{
 const p=await fixture(t);assert.equal(await p.locator('#save-all-plugin-drafts').isDisabled(),true);assert.equal(await p.locator('#stop-plugin-batch').isDisabled(),true);await draft(p,[0,20]);await p.locator('#plugin-only-drafts').check();await p.locator('#plugin-search').fill('/unique-0');assert.equal(await p.locator('#plugin-list h2').count(),1);
 await p.selectOption('#plugin-scope','group');await p.locator('#plugin-group').fill('987654321');const message=await start(p,false);assert.match(message,/全局.*全部 2 个.*其他页和筛选隐藏项.*分别备份.*不回滚/);assert.equal(await p.evaluate(()=>requests.length),0);assert.equal(await p.evaluate(()=>pluginDrafts.size),2);
 await p.evaluate(()=>{window.confirm=()=>{pluginDrafts.get('plugin-000').mode='enabled';return true;};saveAllPluginDrafts();});assert.equal(await p.evaluate(()=>requests.length),0);assert.equal(await p.evaluate(()=>pluginSaving),false);
 await p.evaluate(()=>{window.confirm=()=>{pluginGroup='12345678';return true;};saveAllPluginDrafts();});assert.equal(await p.evaluate(()=>requests.length),0);
});

test('sequential saves update actual states, clear only acknowledged drafts and reflow filtered pages in both scopes',async t=>{
 for(const group of ['0','123456789']){const p=await fixture(t);await p.evaluate(group=>{pluginGroup=group;renderPluginScope();},group);await draft(p,[0,20]);await p.locator('#plugin-only-drafts').check();await p.locator('#plugin-search').fill('/unique-20');await p.evaluate(()=>replies.push({pending:true},{pending:true}));await start(p);await p.waitForFunction(()=>pending.length===1);assert.equal(await p.evaluate(()=>requests.length),1);assert.equal(await p.locator('#reload-plugins').isDisabled(),true);assert.equal(await p.locator('#apply-plugin-scope').isDisabled(),true);assert.equal(await p.locator('#plugin-only-drafts').isDisabled(),true);assert.equal(await p.locator('#discard-plugin-drafts').isDisabled(),true);assert.equal(await p.locator('#save-all-plugin-drafts').isDisabled(),true);assert.equal(await p.locator('#stop-plugin-batch').isDisabled(),false);
 await resolve(p,0,saved(0,group));await p.waitForFunction(()=>pending.length===2);assert.equal(await p.evaluate(()=>pluginDrafts.size),1);assert.equal(await p.evaluate(()=>plugins[0].state.effectiveEnabled),false);assert.match(await p.locator('#plugin-batch-status').textContent(),/已确认成功 1 \/ 2 项/);await p.locator('[data-view=overview]').click();await p.locator('[data-view=plugins]').click();assert.equal(await p.evaluate(()=>pluginSaving),true);
 await resolve(p,1,saved(20,group));await idle(p);assert.equal(await p.evaluate(()=>pluginDrafts.size),0);assert.match(await p.locator('#plugin-batch-status').textContent(),/全部 2 项已确认保存/);assert.equal(await p.locator('#plugin-list h2').count(),0);assert.equal(await p.locator('#stop-plugin-batch').isDisabled(),true);const reqs=await p.evaluate(()=>requests);assert.deepEqual(reqs.map(r=>r.url),['/api/plugin-switch','/api/plugin-switch']);assert.deepEqual(reqs.map(r=>r.body),[0,20].map(i=>({name:saved(i).name,groupId:group,mode:'disabled',revision:'a'.repeat(64)})));
 }
});

test('first conflict or network error stops subsequent items and retains failed drafts, without retrying or undoing successful rules',async t=>{
 for(const failure of [{error:'版本冲突',status:409},{throw:'network interrupted'}]){const p=await fixture(t);await draft(p,[0,1,2]);await p.evaluate(({first,failure})=>replies.push(first,failure),{first:saved(0),failure});await start(p);await idle(p);assert.equal(await p.evaluate(()=>requests.length),2);assert.deepEqual(await p.evaluate(()=>[...pluginDrafts.keys()]),['plugin-001','plugin-002']);assert.equal(await p.evaluate(()=>plugins[0].state.effectiveEnabled),false);assert.equal(await p.evaluate(()=>pluginDrafts.get('plugin-001').revision),'a'.repeat(64));assert.match(await p.locator('#plugin-batch-status').textContent(),/已确认成功 1 \/ 3 项.*余下 2 项.*不会自动重试或回滚/);assert.equal(await p.locator('#save-all-plugin-drafts').isDisabled(),false);
 }
});

test('stop waits for the current item and keeps remaining drafts; stopping on the last successful item reports completion',async t=>{
 for(const ids of [[0,1,2],[0]]){const p=await fixture(t);await draft(p,ids);await p.evaluate(()=>replies.push({pending:true}));await start(p);await p.waitForFunction(()=>pending.length===1);await p.locator('#stop-plugin-batch').click();assert.equal(await p.evaluate(()=>pending[0].signal.aborted),false);assert.equal(await p.evaluate(()=>pluginSaving),true);assert.equal(await p.locator('#stop-plugin-batch').isDisabled(),true);assert.match(await p.locator('#plugin-batch-status').textContent(),/等待当前请求结束/);await resolve(p,0,saved(0));await idle(p);assert.equal(await p.evaluate(()=>requests.length),1);assert.equal(await p.evaluate(()=>pluginDrafts.size),ids.length-1);assert.match(await p.locator('#plugin-batch-status').textContent(),ids.length===1?/全部 1 项已确认保存/:/已停止后续保存.*已确认成功 1 \/ 3 项.*余下 2 项草稿保留/);
 }
});

test('unverifiable result and timeout remain unconfirmed, and invalid preflight cannot start any writes',async t=>{
 const p=await fixture(t);await draft(p,[0,1]);const wrong=saved(0);wrong.groupId='999';await p.evaluate(v=>replies.push(v),wrong);await start(p);await idle(p);assert.equal(await p.evaluate(()=>requests.length),1);assert.equal(await p.evaluate(()=>pluginDrafts.size),2);assert.match(await p.locator('#plugin-error').textContent(),/返回结果无法核对/);assert.equal(await p.evaluate(()=>plugins[0].state.globalMode),'inherit');
 await p.evaluate(()=>{window.originalSetTimeout=window.setTimeout;window.setTimeout=(fn,delay,...args)=>delay===30000?originalSetTimeout(fn,0,...args):originalSetTimeout(fn,delay,...args);replies.push({pending:true});});await start(p);await idle(p);assert.equal(await p.evaluate(()=>requests.length),2);assert.equal(await p.evaluate(()=>pluginDrafts.size),2);assert.match(await p.locator('#plugin-error').textContent(),/超时或中断.*未确认/);await p.evaluate(()=>{window.setTimeout=originalSetTimeout;pluginDrafts.set('missing-plugin',{mode:'disabled',revision:'a'.repeat(64)});saveAllPluginDrafts();});assert.equal(await p.evaluate(()=>requests.length),2);
});

test('expired sessions clear a batch, and old success or old 401 cannot continue it or unlock a new batch',async t=>{
 for(const oldStatus of [200,401]){const p=await fixture(t);await draft(p,[0,1]);await p.evaluate(()=>replies.push({pending:true}));await start(p);await p.waitForFunction(()=>pending.length===1);await p.evaluate(data=>{authenticated=false;csrf='';showAuth();authenticated=true;csrf='new-session';plugins=data;pluginControlAvailable=true;document.getElementById('auth').hidden=true;document.getElementById('workspace').hidden=false;document.getElementById('plugins-view').hidden=false;renderPlugins();},fixtures());assert.equal(await p.locator('#plugin-batch-status').textContent(),'');await draft(p,[2]);await p.evaluate(()=>replies.push({pending:true}));await start(p);await p.waitForFunction(()=>pending.length===2);await resolve(p,0,oldStatus===200?saved(0):{error:'old expiry'},oldStatus);await p.waitForTimeout(50);assert.equal(await p.evaluate(()=>authenticated&&pluginSaving),true);assert.deepEqual(await p.evaluate(()=>[...pluginDrafts.keys()]),['plugin-002']);assert.equal(await p.evaluate(()=>requests.length),2);await resolve(p,1,{error:'current expiry'},401);await p.waitForFunction(()=>!authenticated);assert.equal(await p.evaluate(()=>pluginBatch),null);assert.equal(await p.locator('#plugin-batch-status').textContent(),'');assert.equal(await p.evaluate(()=>pluginDrafts.size),0);
 }
});
