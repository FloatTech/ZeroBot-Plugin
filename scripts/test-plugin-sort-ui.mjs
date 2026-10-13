// Isolated Chrome DOM, synthetic plugins and mocked APIs; no live site or QQ.
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
async function harness(t,data=fixtures()){
  const context=await browser.newContext();t.after(()=>context.close());
  await context.route('**/*',route=>route.abort());
  const page=await context.newPage();await page.setContent(html);await page.addStyleTag({content:css});await page.addScriptTag({content:app});
  await page.evaluate(data=>{
    authenticated=true;csrf='fixture-session';plugins=data;pluginControlAvailable=true;
    window.showOverview=()=>{};window.showLogs=()=>{};window.showBackups=()=>{};
    document.getElementById('auth').hidden=true;document.getElementById('workspace').hidden=false;
    document.getElementById('overview-view').hidden=true;document.getElementById('plugins-view').hidden=false;
    window.requests=[];window.replies=[];window.pending=[];
    window.fetch=async(url,options)=>{
      requests.push({url,body:options.body?JSON.parse(options.body):null});
      const reply=replies.shift();if(!reply)throw Error('Missing API fixture');
      if(reply.pending)return new Promise(resolve=>pending.push(resolve));
      return {ok:!reply.error,status:reply.status||200,json:async()=>reply};
    };
    renderPlugins();
  },data);
  return page;
}
const names=page=>page.locator('#plugin-list h2').allTextContents();
const card=(page,i)=>page.locator(`[data-plugin="plugin-${String(i).padStart(3,'0')}"]`);
const idle=page=>page.waitForFunction(()=>!pluginSaving&&!pluginScopeBusy);
async function confirmAction(page,accept,action){
  const event=page.waitForEvent('dialog'),click=action();const dialog=await event;
  assert.equal(dialog.type(),'confirm');if(accept)await dialog.accept();else await dialog.dismiss();await click;
}

test('sort covers all filtered pages without mutating loaded order; natural names and ties are stable',async t=>{
 const data=fixtures(45,true).reverse(),p=await harness(t,data);
 await p.locator('#plugin-next').click();await p.selectOption('#plugin-sort','name');assert.equal(await p.evaluate(()=>pluginPage),1);
 const collected=[];for(let i=0;i<3;i++){collected.push(...await names(p));if(i<2)await p.locator('#plugin-next').click();}
 assert.deepEqual(collected,fixtures(45).map(p=>p.name));assert.equal(new Set(collected).size,45);assert.deepEqual(await p.evaluate(()=>plugins.map(p=>p.name)),data.map(p=>p.name));
 await p.locator('#plugin-search').fill('/unique-3');assert.deepEqual(await names(p),['plugin-003',...fixtures(10).map((_,i)=>'plugin-0'+(30+i))]);
 await p.locator('#clear-plugin-filter').click();assert.equal(await p.locator('#plugin-sort').inputValue(),'name');await p.selectOption('#plugin-sort','default');assert.deepEqual(await names(p),data.slice(0,20).map(p=>p.name));
 await p.evaluate(()=>{plugins=['plugin-10','plugin-2','plugin-1','Plugin-1','<img src=x>'].map(name=>({name,brief:'literal',usage:'usage'}));renderPlugins();});await p.selectOption('#plugin-sort','name');
 const natural=await names(p);assert.ok(natural.indexOf('plugin-2')<natural.indexOf('plugin-10'));assert.ok(natural.indexOf('plugin-1')<natural.indexOf('Plugin-1'));assert.equal(await p.locator('#plugin-list img').count(),0);
 assert.equal(await p.evaluate(()=>requests.length),0);
});

test('actual state priorities keep unknown last and ignore drafts; draft priority updates without losing hidden selections',async t=>{
 const data=fixtures(30,true).reverse(),p=await harness(t,data);await p.selectOption('#plugin-page-size','50');
 for(const order of ['enabled','disabled']){
  await p.selectOption('#plugin-sort',order);const enabled=data.filter(p=>p.state?.effectiveEnabled===true).map(p=>p.name),disabled=data.filter(p=>p.state?.effectiveEnabled===false).map(p=>p.name),unknown=data.filter(p=>!p.state).map(p=>p.name);
  assert.deepEqual(await names(p),[...(order==='enabled'?enabled:disabled),...(order==='enabled'?disabled:enabled),...unknown]);
 }
 await card(p,0).locator('select').selectOption('disabled');assert.equal((await names(p))[0],'plugin-028');await card(p,1).locator('select').selectOption('enabled');
 await p.selectOption('#plugin-sort','drafts');assert.deepEqual((await names(p)).slice(0,2),['plugin-001','plugin-000']);
 await card(p,3).locator('select').selectOption('disabled');assert.deepEqual((await names(p)).slice(0,3),['plugin-003','plugin-001','plugin-000']);
 await card(p,3).locator('select').selectOption('inherit');assert.deepEqual((await names(p)).slice(0,2),['plugin-001','plugin-000']);
 await p.locator('#plugin-search').fill('unique-29');assert.equal(await p.evaluate(()=>pluginDrafts.size),2);await p.locator('#plugin-only-drafts').check();assert.equal((await names(p)).length,0);
 await p.locator('#clear-plugin-filter').click();assert.equal(await p.locator('#plugin-sort').inputValue(),'drafts');assert.deepEqual((await names(p)).slice(0,2),['plugin-001','plugin-000']);
 await card(p,1).locator('.text-button').click();await p.evaluate(()=>{pluginSort='name';renderPlugins()});assert.equal(await p.locator('#usage-title').textContent(),'plugin-001');await p.locator('#close-dialog').click();
 assert.equal(await p.evaluate(()=>requests.length),0);
});

test('sort survives navigation, refresh and new scopes, while failures, busy controls and invalid choices preserve current state',async t=>{
 const p=await harness(t,fixtures(45,true));await p.selectOption('#plugin-sort','disabled');await p.locator('#plugin-next').click();
 await p.locator('[data-view=overview]').click();await p.locator('[data-view=plugins]').click();assert.equal(await p.evaluate(()=>pluginPage),2);assert.equal(await p.locator('#plugin-sort').inputValue(),'disabled');
 await p.evaluate(()=>replies.push({error:'fixture failure',status:503}));await p.locator('#reload-plugins').click();await idle(p);assert.equal(await p.evaluate(()=>pluginPage),2);assert.equal(await p.locator('#plugin-sort').inputValue(),'disabled');
 await p.evaluate(data=>replies.push({plugins:data,controlAvailable:true}),fixtures(45,true).reverse());await p.locator('#reload-plugins').click();await idle(p);assert.equal(await p.evaluate(()=>pluginPage),2);assert.equal(await p.locator('#plugin-sort').inputValue(),'disabled');
 await p.selectOption('#plugin-scope','group');await p.locator('#plugin-group').fill('12345678');await p.evaluate(data=>replies.push({plugins:data,controlAvailable:true}),fixtures(7,true));await p.locator('#apply-plugin-scope').click();await idle(p);assert.equal(await p.evaluate(()=>pluginPage),1);assert.equal(await p.locator('#plugin-sort').inputValue(),'disabled');assert.deepEqual(await names(p),['plugin-001','plugin-004','plugin-000','plugin-003','plugin-006','plugin-002','plugin-005']);
 for(const busy of ['pluginSaving','pluginScopeBusy']){await p.evaluate(busy=>{window.eval(busy+'=true');renderPluginBusy()},busy);assert.equal(await p.locator('#plugin-sort').isDisabled(),true);await p.evaluate(()=>{const s=document.getElementById('plugin-sort');s.value='name';s.dispatchEvent(new Event('change'))});assert.equal(await p.locator('#plugin-sort').inputValue(),'disabled');await p.evaluate(busy=>{window.eval(busy+'=false');renderPluginBusy()},busy);}
 await p.evaluate(()=>{const s=document.getElementById('plugin-sort');s.value='invalid';s.dispatchEvent(new Event('change'))});assert.equal(await p.locator('#plugin-sort').inputValue(),'disabled');
});

test('single save addresses the sorted card; sequential save keeps loaded order even when sorted/filtered differently',async t=>{
 const data=fixtures(4).reverse(),p=await harness(t,data);await p.selectOption('#plugin-sort','name');await card(p,0).locator('select').selectOption('disabled');
 await p.evaluate(()=>replies.push({state:{effectiveEnabled:false,defaultEnabled:true,globalMode:'disabled',groupMode:'inherit',source:'global',revision:'b'.repeat(64)}}));await card(p,0).locator('.plugin-controls button').click();await idle(p);assert.equal(await p.evaluate(()=>requests[0].body.name),'plugin-000');assert.equal(await p.evaluate(()=>plugins.find(p=>p.name==='plugin-000').state.globalMode),'disabled');
 for(const i of [1,3])await card(p,i).locator('select').selectOption('disabled');await p.locator('#plugin-search').fill('/unique-1');
 await p.evaluate(()=>{for(const name of ['plugin-003','plugin-001'])replies.push({name,groupId:'0',state:{effectiveEnabled:false,defaultEnabled:true,globalMode:'disabled',groupMode:'inherit',source:'global',revision:'b'.repeat(64)}})});
 await confirmAction(p,true,()=>p.locator('#save-all-plugin-drafts').click());await idle(p);assert.deepEqual(await p.evaluate(()=>requests.map(r=>r.body.name)),['plugin-000','plugin-003','plugin-001']);assert.equal(await p.evaluate(()=>pluginDrafts.size),0);assert.equal(await p.locator('#plugin-sort').inputValue(),'name');assert.equal(await p.locator('#plugin-sort').isDisabled(),false);
});

test('logout/expiry reset sorting and late old-session replies cannot change the new sorted list',async t=>{
 for(const status of [200,401]){
  const p=await harness(t);await p.selectOption('#plugin-sort','disabled');await p.evaluate(()=>replies.push({pending:true}));await p.locator('#reload-plugins').click();await p.waitForFunction(()=>pending.length===1);assert.equal(await p.locator('#plugin-sort').isDisabled(),true);
  await p.evaluate(data=>{authenticated=false;csrf='';showAuth();window.resetOrder=pluginSort;authenticated=true;csrf='new-sort-session';plugins=data;pluginControlAvailable=true;document.getElementById('auth').hidden=true;document.getElementById('workspace').hidden=false;renderPluginBusy();renderPlugins()},fixtures(3).reverse());assert.equal(await p.evaluate(()=>resetOrder),'default');
  await p.selectOption('#plugin-sort','name');await p.evaluate(({status,data})=>pending[0]({ok:status===200,status,json:async()=>status===200?{plugins:data,controlAvailable:true}:{error:'old expiry'}}),{status,data:fixtures(108)});await p.waitForTimeout(40);
  assert.equal(await p.evaluate(()=>authenticated),true);assert.equal(await p.locator('#plugin-sort').inputValue(),'name');assert.deepEqual(await names(p),['plugin-000','plugin-001','plugin-002']);
  await p.evaluate(()=>replies.push({error:'current expiry',status:401}));await p.locator('#reload-plugins').click();await p.waitForFunction(()=>!authenticated);assert.equal(await p.locator('#plugin-sort').inputValue(),'default');assert.equal(await p.locator('#plugin-sort').isDisabled(),true);assert.equal((await names(p)).length,0);
 }
});

