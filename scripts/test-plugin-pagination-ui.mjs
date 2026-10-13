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

test('108 plugins paginate without loss or duplication, then filter all records before paging',async t=>{
  const page=await harness(t,fixtures(108,true));
  assert.equal(await page.locator('#plugin-page-size').inputValue(),'20');
  assert.equal(await page.locator('#plugin-prev').isDisabled(),true);
  const collected=[];
  for(let i=0;i<6;i++){
    collected.push(...await names(page));assert.equal((await names(page)).length,i===5?8:20);
    if(i<5)await page.locator('#plugin-next').click();
  }
  assert.deepEqual(collected,fixtures().map(p=>p.name));assert.equal(new Set(collected).size,108);
  assert.equal(await page.locator('#plugin-next').isDisabled(),true);
  assert.match(await page.locator('#plugin-page-status').textContent(),/101～108 \/ 108/);
  await page.selectOption('#plugin-page-size','10');assert.equal(await page.evaluate(()=>pluginPage),1);assert.equal((await names(page)).length,10);
  await page.selectOption('#plugin-page-size','50');await page.locator('#plugin-next').click();assert.equal((await names(page)).length,50);
  await page.locator('#plugin-search').fill('/unique-107');assert.deepEqual(await names(page),['plugin-107']);assert.equal(await page.evaluate(()=>pluginPage),1);
  await page.locator('#clear-plugin-filter').click();await page.selectOption('#plugin-state-filter','unknown');assert.equal((await names(page)).length,36);
  assert.match(await page.locator('#search-count').textContent(),/36 \/ 108 个插件；启用 36，禁用 36，未知 36/);
  await page.locator('#plugin-search').fill('no-match');assert.equal((await names(page)).length,0);assert.equal(await page.locator('#plugin-next').isDisabled(),true);assert.equal(await page.locator('#plugin-prev').isDisabled(),true);
  await page.locator('#clear-plugin-filter').click();assert.equal(await page.evaluate(()=>pluginPage),1);
  assert.equal(await page.evaluate(()=>requests.length),0);
});

test('hidden drafts survive pagination, page size, filtering and navigation; usage retains its plugin',async t=>{
  const page=await harness(t);
  await card(page,0).locator('select').selectOption('disabled');await page.locator('#plugin-next').click();
  await card(page,20).locator('select').selectOption('enabled');assert.match(await page.locator('#plugin-draft-status').textContent(),/2 个插件/);
  await page.locator('[data-view="overview"]').click();await page.locator('[data-view="plugins"]').click();
  assert.equal(await page.evaluate(()=>pluginPage),2);assert.equal(await card(page,20).locator('select').inputValue(),'enabled');
  await page.locator('#plugin-prev').click();assert.equal(await card(page,0).locator('select').inputValue(),'disabled');assert.equal(await card(page,0).locator('.plugin-draft-note').isVisible(),true);
  await page.selectOption('#plugin-page-size','50');assert.equal(await card(page,20).locator('select').inputValue(),'enabled');
  await page.locator('#plugin-search').fill('/unique-20');assert.equal((await names(page)).length,1);assert.match(await page.locator('#plugin-draft-status').textContent(),/2 个插件/);
  await page.locator('[data-view="overview"]').click();await page.locator('[data-view="plugins"]').click();assert.equal(await card(page,20).locator('select').inputValue(),'enabled');
  await page.locator('#clear-plugin-filter').click();await card(page,20).locator('.text-button').click();
  // A modal blocks real underlying clicks; re-rendering alone must not retarget it.
  await page.evaluate(()=>changePluginPage(1));assert.equal(await page.locator('#usage-title').textContent(),'plugin-020');assert.equal(await page.locator('#usage-text').textContent(),'/unique-20');
  await page.locator('#close-dialog').click();await page.locator('#plugin-prev').click();
  assert.equal(await card(page,0).locator('select').inputValue(),'disabled');assert.equal(await card(page,20).locator('select').inputValue(),'enabled');
  assert.equal(await page.evaluate(()=>requests.length),0);
});

test('busy save freezes pages, filtered save clamps shrinking results and preserves other drafts on conflict',async t=>{
  const page=await harness(t,fixtures(21));
  await card(page,0).locator('select').selectOption('disabled');await page.selectOption('#plugin-state-filter','enabled');await page.locator('#plugin-next').click();
  await card(page,20).locator('select').selectOption('disabled');await page.evaluate(()=>replies.push({pending:true}));
  await card(page,20).locator('.plugin-controls button').click();await page.waitForFunction(()=>pending.length===1);
  for(const id of ['plugin-prev','plugin-next','plugin-page-size'])assert.equal(await page.locator('#'+id).isDisabled(),true);
  await page.evaluate(()=>changePluginPage(-1));assert.equal(await page.evaluate(()=>pluginPage),2);
  await page.evaluate(saved=>pending[0]({ok:true,status:200,json:async()=>({state:saved})}),state(false,'disabled'));await idle(page);
  assert.equal(await page.evaluate(()=>pluginPage),1);assert.equal((await names(page)).length,20);assert.deepEqual(await page.evaluate(()=>[...pluginDrafts.keys()]),['plugin-000']);
  await page.evaluate(()=>replies.push({error:'版本冲突',status:409}));await card(page,0).locator('.plugin-controls button').click();await idle(page);
  assert.equal(await card(page,0).locator('select').inputValue(),'disabled');assert.equal(await page.evaluate(()=>pluginDrafts.size),1);
  assert.deepEqual(await page.evaluate(()=>requests.map(r=>r.body.name)),['plugin-020','plugin-000']);
  assert.equal(await page.evaluate(()=>requests[0].body.groupId),'0');
});

test('refresh preserves page or clamps it, failed scope retains page and drafts, successful scope resets',async t=>{
  const page=await harness(t);for(let i=0;i<5;i++)await page.locator('#plugin-next').click();
  await card(page,100).locator('select').selectOption('disabled');
  await confirmAction(page,false,()=>page.locator('#reload-plugins').click());assert.equal(await page.evaluate(()=>requests.length),0);
  await page.evaluate(()=>replies.push({error:'读取失败',status:500}));await confirmAction(page,true,()=>page.locator('#reload-plugins').click());await idle(page);
  assert.equal(await page.evaluate(()=>pluginPage),6);assert.equal(await card(page,100).locator('select').inputValue(),'disabled');
  await page.selectOption('#plugin-scope','group');await page.locator('#plugin-group').fill('234567890');
  await page.evaluate(()=>replies.push({error:'范围读取失败',status:500}));await confirmAction(page,true,()=>page.locator('#apply-plugin-scope').click());await idle(page);
  assert.equal(await page.evaluate(()=>pluginGroup),'0');assert.equal(await page.evaluate(()=>pluginPage),6);assert.equal(await page.evaluate(()=>pluginDrafts.size),1);
  await page.evaluate(data=>replies.push({plugins:data,controlAvailable:true}),fixtures());await confirmAction(page,true,()=>page.locator('#reload-plugins').click());await idle(page);assert.equal(await page.evaluate(()=>pluginPage),6);
  await page.evaluate(data=>replies.push({plugins:data,controlAvailable:true}),fixtures(25));await page.locator('#reload-plugins').click();await idle(page);
  assert.equal(await page.evaluate(()=>pluginPage),2);assert.equal((await names(page)).length,5);
  await page.selectOption('#plugin-scope','group');await page.locator('#plugin-group').fill('234567890');
  await page.evaluate(data=>replies.push({plugins:data,controlAvailable:true}),fixtures());await page.locator('#apply-plugin-scope').click();await idle(page);
  assert.equal(await page.evaluate(()=>pluginGroup),'234567890');assert.equal(await page.evaluate(()=>pluginPage),1);
});

test('logout and expiry reset pages and cache, and late responses cannot replace a new paged session',async t=>{
  const page=await harness(t);await page.selectOption('#plugin-page-size','10');await page.locator('#plugin-next').click();
  await page.evaluate(()=>replies.push({pending:true}));await page.locator('#reload-plugins').click();await page.waitForFunction(()=>pending.length===1);
  await page.evaluate(data=>{
    authenticated=false;csrf='';showAuth();
    authenticated=true;csrf='new-session';plugins=data;pluginControlAvailable=true;
    document.getElementById('auth').hidden=true;document.getElementById('workspace').hidden=false;renderPlugins();
  },fixtures(65));
  assert.equal(await page.locator('#plugin-page-size').inputValue(),'20');assert.equal(await page.evaluate(()=>pluginPage),1);
  await page.locator('#plugin-next').click();await page.evaluate(()=>replies.push({pending:true}));await page.locator('#reload-plugins').click();await page.waitForFunction(()=>pending.length===2);
  await page.evaluate(data=>pending[0]({ok:true,status:200,json:async()=>({plugins:data,controlAvailable:true})}),fixtures(1));await page.waitForTimeout(30);
  assert.equal(await page.evaluate(()=>pluginScopeBusy),true);assert.equal(await page.evaluate(()=>plugins.length),65);assert.equal(await page.evaluate(()=>pluginPage),2);
  await page.evaluate(data=>pending[1]({ok:true,status:200,json:async()=>({plugins:data,controlAvailable:true})}),fixtures(65));await idle(page);
  await page.evaluate(()=>replies.push({error:'会话过期',status:401}));await page.locator('#reload-plugins').click();await page.waitForFunction(()=>!authenticated);
  assert.equal(await page.evaluate(()=>pluginPage),1);assert.equal(await page.locator('#plugin-page-size').inputValue(),'20');assert.equal(await page.evaluate(()=>plugins.length),0);assert.equal(await page.locator('#plugin-list').textContent(),'');
  await page.evaluate(data=>{authenticated=true;csrf='last-session';plugins=data;document.getElementById('auth').hidden=true;document.getElementById('workspace').hidden=false;renderPlugins();replies.push({});},fixtures(30));
  await page.selectOption('#plugin-page-size','10');await page.locator('#plugin-next').click();await page.locator('#logout').click();await page.waitForFunction(()=>!authenticated);
  assert.equal(await page.evaluate(()=>pluginPage),1);assert.equal(await page.evaluate(()=>pluginPageSize),20);assert.equal(await page.evaluate(()=>pluginDrafts.size),0);
});
