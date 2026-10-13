// Offline Chrome synthetic plugin metadata and mocked APIs; no live site, QQ or configuration mutations.
import assert from 'node:assert/strict';
import fs from 'node:fs/promises';
import path from 'node:path';
import {createRequire} from 'node:module';
import test, {after} from 'node:test';
const require=createRequire(import.meta.url);
const modules=process.env.LOCALADMIN_NODE_MODULES || path.join(process.env.USERPROFILE,'.cache/codex-runtimes/codex-primary-runtime/dependencies/node/node_modules');
const {chromium}=require(path.join(modules,'playwright'));
const html=(await fs.readFile('internal/localadmin/assets/index.html','utf8')).replace(/<script\b[^>]*>[\s\S]*?<\/script>/g,'').replace(/<link\b[^>]*>/g,'');
const app=await fs.readFile('internal/localadmin/assets/app.js','utf8');
const css=await fs.readFile('internal/localadmin/assets/style.css','utf8')+'\n'+await fs.readFile('internal/localadmin/assets/appearance.css','utf8');
const browser=await chromium.launch({channel:'chrome',headless:true});
after(()=>browser.close());
const state=(enabled=true,mode='inherit')=>({effectiveEnabled:enabled,defaultEnabled:true,globalMode:mode,groupMode:'inherit',source:'default',revision:'a'.repeat(64)});
const fixtures=(n=108,mixed=false)=>Array.from({length:n},(_,i)=>({name:'plugin-'+String(i).padStart(3,'0'),brief:'功能 '+i,usage:'/unique-'+i,defaultEnabledAtStartup:true,...(mixed&&i%3===2?{}:{state:state(!mixed||i%3===0)})}));
async function fixture(t,data=fixtures(65,true)) {
 const context=await browser.newContext({acceptDownloads:true});t.after(()=>context.close());await context.route('**/*',r=>r.abort());
 const p=await context.newPage();await p.setContent(html);await p.addStyleTag({content:css});await p.addScriptTag({content:app});
 await p.evaluate(data=>{authenticated=true;csrf='catalog-session';window.showOverview=()=>{};window.showLogs=()=>{};window.showBackups=()=>{};document.getElementById('auth').hidden=true;document.getElementById('workspace').hidden=false;document.getElementById('overview-view').hidden=true;document.getElementById('plugins-view').hidden=false;window.data=data;window.requests=[];window.replies=[];window.pending=[];window.revoked=[];const revoke=URL.revokeObjectURL.bind(URL);URL.revokeObjectURL=u=>{revoked.push(u);revoke(u)};window.fetch=async(url,options)=>{requests.push({url,method:options.method||'GET',body:options.body?JSON.parse(options.body):null});const reply=replies.shift();if(!reply)throw Error('Missing API fixture');if(reply.pending)return new Promise(resolve=>pending.push(resolve));return {ok:!reply.error,status:reply.status||200,json:async()=>reply};};},data);
 return p;
}
async function read(p,group='0') {await p.evaluate(async group=>{replies.push({plugins:data,controlAvailable:true,groupId:group});await loadPluginList(group,true);renderPluginBusy()},group);}
async function download(p) {const [d]=await Promise.all([p.waitForEvent('download'),p.locator('#download-plugin-catalog').click()]);const text=await fs.readFile(await d.path(),'utf8');return {d,text,value:JSON.parse(text)};}
const idle=p=>p.waitForFunction(()=>!pluginSaving&&!pluginScopeBusy&&!pluginCatalogLoading);
const card=(p,i)=>p.locator(`[data-plugin="plugin-${String(i).padStart(3,'0')}"]`);

test('all matching plugin pages export in selected order with explicit fields, readable Unicode and recycled URL',async t=>{
 const data=fixtures(65,true).reverse().map(p=>({...p,secret:'fixture-private',config:{token:'fixture-private'},state:p.state?{...p.state,extra:'fixture-private'}:undefined})),p=await fixture(t,data);await read(p);
 await p.selectOption('#plugin-sort','name');await p.locator('#plugin-next').click();const out=await download(p);
 assert.match(out.d.suggestedFilename(),/^zerobot-plugin-state-catalog-\d{8}T\d{9}Z\.json$/);assert.equal(out.value.format,'zerobot-plugin-state-catalog');assert.equal(out.value.version,1);assert.deepEqual(out.value.scope,{kind:'global',groupId:null});assert.deepEqual(out.value.source,{loaded:65,matched:65,unsaved:0,lastReadFailed:false,liveSnapshot:false});assert.deepEqual(out.value.filters,{state:'all',keywordApplied:false,onlyUnsaved:false,sort:'name'});assert.deepEqual(out.value.plugins.map(p=>p.name),fixtures(65).map(p=>p.name));assert.equal(await p.evaluate(()=>pluginPage),2);
 assert.deepEqual(Object.keys(out.value.plugins[0]),['name','defaultEnabledAtStartup','effectiveState','defaultEnabled','globalMode','groupMode','source','hasUnsavedChanges']);assert.doesNotMatch(out.text,/fixture-private|revision|"(?:config|token|usage|brief|extra)"\s*:/);assert.equal(await p.evaluate(()=>requests.length),1);assert.equal(await p.locator('a[download]').count(),0);await p.waitForFunction(()=>revoked.length===1);assert.ok(Number.isFinite(Date.parse(out.value.readAt)));assert.ok(Date.parse(out.value.exportedAt)>=Date.parse(out.value.readAt));
 await p.locator('#plugin-search').fill('/unique-3');await p.selectOption('#plugin-state-filter','enabled');const filtered=await download(p);assert.deepEqual(filtered.value.plugins.map(p=>p.name),['plugin-003','plugin-030','plugin-033','plugin-036','plugin-039']);assert.equal(filtered.value.filters.keywordApplied,true);assert.doesNotMatch(filtered.text,/unique-3/);assert.equal(await p.evaluate(()=>requests.length),1);
});

test('unknown and empty results remain explicit and do not infer live state from defaults; export excludes registry text',async t=>{
 const data=[{name:'中文插件 <img src=x>',brief:'fixture-private-brief',usage:'fixture-private-usage',defaultEnabledAtStartup:true},{name:'malformed',state:{effectiveEnabled:'true',defaultEnabled:true,globalMode:'enabled',source:'default',revision:'fixture-private'}}],p=await fixture(t,data);await read(p);await p.selectOption('#plugin-state-filter','unknown');const out=await download(p);
 assert.equal(out.value.plugins.length,2);for(const item of out.value.plugins){assert.equal(item.effectiveState,'unknown');for(const field of ['defaultEnabled','globalMode','groupMode','source'])assert.equal(item[field],null)}assert.match(out.text,/中文插件/);assert.doesNotMatch(out.text,/fixture-private/);assert.equal(await p.locator('#plugin-list img').count(),0);
 await p.locator('#plugin-search').fill('unmatched');const empty=await download(p);assert.equal(empty.value.source.matched,0);assert.deepEqual(empty.value.plugins,[]);assert.equal(await p.locator('#download-plugin-catalog').isEnabled(),true);
 await p.evaluate(()=>data=[]);await read(p);const noPlugins=await download(p);assert.equal(noPlugins.value.source.loaded,0);assert.deepEqual(noPlugins.value.plugins,[]);
});

test('group scope uses loaded target; drafts are flags only and export preserves hidden drafts, page and usage dialog',async t=>{
 const p=await fixture(t,fixtures(45).map(p=>({...p,state:{...state(true,'enabled'),groupMode:'disabled',source:'global'}})));await read(p,'42');await p.locator('#plugin-group').fill('999');await card(p,0).locator('select').selectOption('inherit');await card(p,1).locator('select').selectOption('enabled');await card(p,0).locator('.text-button').click();assert.equal(await p.locator('#usage-dialog').getAttribute('open'),'');await p.locator('#close-dialog').click();
 await p.locator('#plugin-search').fill('plugin-000');const out=await download(p);assert.deepEqual(out.value.scope,{kind:'group',groupId:'42'});assert.equal(out.value.source.unsaved,2);assert.deepEqual(out.value.plugins[0],{name:'plugin-000',defaultEnabledAtStartup:true,effectiveState:'enabled',defaultEnabled:true,globalMode:'enabled',groupMode:'disabled',source:'global',hasUnsavedChanges:true});assert.doesNotMatch(out.text,/999|"mode"\s*:/);assert.equal(await p.evaluate(()=>pluginDrafts.size),2);
 await p.locator('#clear-plugin-filter').click();await p.locator('#plugin-only-drafts').check();const drafts=await download(p);assert.deepEqual(drafts.value.plugins.map(p=>p.name),['plugin-000','plugin-001']);assert.equal(drafts.value.filters.onlyUnsaved,true);assert.equal(await p.evaluate(()=>requests.length),1);
 await p.locator('#clear-plugin-filter').click();await p.locator('#plugin-next').click();await p.evaluate(()=>openPluginUsage(plugins[0]));await p.evaluate(()=>downloadPluginCatalog());assert.equal(await p.locator('#usage-dialog').getAttribute('open'),'');assert.equal(await p.evaluate(()=>pluginPage),2);assert.equal(await p.evaluate(()=>pluginDrafts.size),2);
});

test('single and batch saves update only returned states and last save time; reload refreshes source and clears save timestamp',async t=>{
 const p=await fixture(t,fixtures(3));await read(p);const readAt=await p.evaluate(()=>pluginCatalogReadAt);await card(p,0).locator('select').selectOption('disabled');await p.evaluate(()=>replies.push({state:{...plugins[0].state,effectiveEnabled:false,globalMode:'disabled'}}));await card(p,0).locator('button.secondary').filter({hasText:'保存开关'}).click();await idle(p);
 const saved=await download(p);assert.equal(saved.value.readAt,readAt);assert.ok(Number.isFinite(Date.parse(saved.value.lastSuccessfulSwitchSaveAt)));assert.equal(saved.value.plugins[0].effectiveState,'disabled');assert.equal(saved.value.plugins[1].effectiveState,'enabled');assert.equal(saved.value.plugins[0].hasUnsavedChanges,false);
 await card(p,1).locator('select').selectOption('disabled');await p.evaluate(()=>{window.confirm=()=>true;replies.push({name:plugins[1].name,groupId:'0',state:{...plugins[1].state,effectiveEnabled:false,globalMode:'disabled'}})});await p.locator('#save-all-plugin-drafts').click();await idle(p);const batch=await download(p);assert.equal(batch.value.plugins[1].effectiveState,'disabled');assert.equal(batch.value.source.unsaved,0);
 await read(p);assert.equal(await p.evaluate(()=>pluginCatalogStateAt),null);assert.equal(await p.evaluate(()=>requests.length),4);
});

test('unread, busy, inactive and hidden downloads are blocked; failed reads preserve loaded scope and mark stale source',async t=>{
 const p=await fixture(t);assert.equal(await p.locator('#download-plugin-catalog').isDisabled(),true);await p.evaluate(()=>downloadPluginCatalog());assert.equal(await p.locator('a[download]').count(),0);await read(p,'42');await p.evaluate(()=>replies.push({pending:true}));await p.locator('#reload-plugins').click();await p.waitForFunction(()=>pending.length===1);assert.equal(await p.locator('#download-plugin-catalog').isDisabled(),true);await p.evaluate(()=>downloadPluginCatalog());
 await p.evaluate(()=>pending[0]({ok:false,status:503,json:async()=>({error:'fixture failure'})}));await idle(p);assert.equal(await p.locator('#download-plugin-catalog').isEnabled(),true);const stale=await download(p);assert.equal(stale.value.source.lastReadFailed,true);assert.equal(stale.value.scope.groupId,'42');assert.match(await p.locator('#plugin-catalog-source').textContent(),/读取失败.*旧结果/);
 await p.locator('[data-view=overview]').click();assert.equal(await p.locator('#download-plugin-catalog').isDisabled(),true);await p.evaluate(()=>downloadPluginCatalog());await p.locator('[data-view=plugins]').click();assert.equal(await p.locator('#download-plugin-catalog').isEnabled(),true);
 await p.evaluate(()=>{Object.defineProperty(document,'hidden',{configurable:true,value:true});document.dispatchEvent(new Event('visibilitychange'));downloadPluginCatalog()});assert.equal(await p.locator('#download-plugin-catalog').isDisabled(),true);await p.evaluate(()=>{Object.defineProperty(document,'hidden',{configurable:true,value:false});document.dispatchEvent(new Event('visibilitychange'))});assert.equal(await p.locator('#download-plugin-catalog').isEnabled(),true);
 await p.evaluate(()=>{pluginSaving=true;renderPluginCatalogControls();downloadPluginCatalog()});assert.equal(await p.locator('#download-plugin-catalog').isDisabled(),true);await p.evaluate(()=>{pluginSaving=false;renderPluginCatalogControls()});await read(p,'42');assert.equal(await p.evaluate(()=>pluginCatalogReadFailed),false);
});

test('logout, expiry and restore invalidation clear source; stale reads cannot refill or unlock a new loading scope',async t=>{
 for(const status of [200,401]){const p=await fixture(t);await read(p);await p.evaluate(()=>replies.push({pending:true}));await p.locator('#reload-plugins').click();await p.waitForFunction(()=>pending.length===1);await p.evaluate(()=>{authenticated=false;csrf='';showAuth()});assert.equal(await p.evaluate(()=>pluginCatalogReadAt),null);assert.equal(await p.locator('#download-plugin-catalog').isDisabled(),true);
 await p.evaluate(()=>{authenticated=true;csrf='fresh-session';document.getElementById('auth').hidden=true;document.getElementById('workspace').hidden=false;document.getElementById('plugins-view').hidden=false;replies.push({pending:true});window.newRead=loadPluginList('77')});await p.waitForFunction(()=>pending.length===2);
 await p.evaluate(({status,data})=>pending[0]({ok:status===200,status,json:async()=>status===200?{plugins:data,controlAvailable:true}:{error:'old expiry'}}),{status,data:fixtures(108)});await p.waitForTimeout(30);assert.equal(await p.evaluate(()=>authenticated),true);assert.equal(await p.locator('#download-plugin-catalog').isDisabled(),true);assert.equal(await p.evaluate(()=>pluginCatalogReadAt),null);
 await p.evaluate(()=>pending[1]({ok:true,status:200,json:async()=>({plugins:data.slice(0,2),controlAvailable:true,groupId:'77'})}));await idle(p);const fresh=await download(p);assert.equal(fresh.value.scope.groupId,'77');assert.equal(fresh.value.source.loaded,2);
 await p.evaluate(()=>{plugins=[];pluginDrafts.clear();pluginListRequest++;resetPluginCatalog();renderPlugins()});assert.equal(await p.locator('#download-plugin-catalog').isDisabled(),true);assert.equal(await p.evaluate(()=>pluginCatalogReadAt),null);await read(p);await p.evaluate(()=>replies.push({error:'expiry',status:401}));await p.locator('#reload-plugins').click();await p.waitForFunction(()=>!authenticated);assert.equal(await p.evaluate(()=>pluginCatalogReadAt),null);assert.equal(await p.locator('#plugin-catalog-status').textContent(),'');}
});

test('download failures keep state and allow retry without extra requests or leaked temporary links',async t=>{
 const p=await fixture(t);await read(p);await p.evaluate(()=>{window.originalCreate=URL.createObjectURL;URL.createObjectURL=()=>{throw Error('fixture download blocked')}});await p.locator('#download-plugin-catalog').click();assert.match(await p.locator('#plugin-catalog-status').textContent(),/下载未完成/);assert.equal(await p.locator('#download-plugin-catalog').isEnabled(),true);assert.equal(await p.locator('a[download]').count(),0);await p.evaluate(()=>URL.createObjectURL=originalCreate);const out=await download(p);assert.equal(out.value.plugins.length,65);assert.equal(await p.evaluate(()=>requests.length),1);
});
