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

test('pending view combines all-page drafts with search and actual state, preserving hidden selections without requests',async t=>{
  const page=await harness(t);
  await page.locator('#plugin-only-drafts').check();assert.equal((await names(page)).length,0);assert.match(await page.locator('#plugin-empty').textContent(),/当前范围没有未保存/);
  await page.locator('#plugin-only-drafts').uncheck();await card(page,0).locator('select').selectOption('disabled');
  await page.locator('#plugin-next').click();await card(page,20).locator('select').selectOption('enabled');
  await page.locator('#plugin-only-drafts').check();assert.equal(await page.evaluate(()=>pluginPage),1);assert.deepEqual(await names(page),['plugin-000','plugin-020']);
  assert.match(await page.locator('#search-count').textContent(),/2 \/ 108 个插件；启用 108，禁用 0/);
  await page.selectOption('#plugin-state-filter','disabled');assert.equal((await names(page)).length,0);assert.match(await page.locator('#plugin-empty').textContent(),/其他草稿仍保留/);
  assert.match(await page.locator('#plugin-draft-status').textContent(),/2 个插件/);
  await page.selectOption('#plugin-state-filter','all');await page.locator('#plugin-search').fill('/unique-20');assert.deepEqual(await names(page),['plugin-020']);
  await page.locator('[data-view="overview"]').click();await page.locator('[data-view="plugins"]').click();assert.equal(await page.locator('#plugin-only-drafts').isChecked(),true);
  await page.locator('#clear-plugin-filter').click();assert.equal(await page.locator('#plugin-only-drafts').isChecked(),false);assert.equal(await page.evaluate(()=>pluginDrafts.size),2);
  await page.locator('#plugin-next').click();assert.equal(await card(page,20).locator('select').inputValue(),'enabled');assert.equal(await page.evaluate(()=>requests.length),0);
});

test('pending results paginate completely and reverting a final-page draft removes it and clamps the page',async t=>{
  const page=await harness(t);
  await page.evaluate(()=>{for(const p of plugins.slice(0,41))pluginDrafts.set(p.name,{mode:'disabled',revision:p.state.revision});renderPlugins();});
  await page.locator('#plugin-only-drafts').check();const all=[];
  for(let i=0;i<3;i++){all.push(...await names(page));if(i<2)await page.locator('#plugin-next').click();}
  assert.deepEqual(all,fixtures(41).map(p=>p.name));assert.equal(new Set(all).size,41);
  await card(page,40).locator('select').selectOption('inherit');assert.equal(await page.evaluate(()=>pluginPage),2);assert.equal((await names(page)).length,20);
  assert.equal(await page.evaluate(()=>pluginDrafts.size),40);assert.equal(await card(page,40).count(),0);
  assert.match(await page.locator('#plugin-page-status').textContent(),/21～40 \/ 40/);assert.equal(await page.locator('#plugin-next').isDisabled(),true);
  await page.locator('#clear-plugin-filter').click();await page.locator('#plugin-next').click();await page.locator('#plugin-next').click();assert.equal(await card(page,40).locator('select').inputValue(),'inherit');
  assert.equal(await page.evaluate(()=>requests.length),0);
});

test('cancel retains visible and hidden drafts; confirmed local discard restores selections but never undoes a saved rule',async t=>{
  const data=fixtures(21);data[1].state=state(false,'disabled');const page=await harness(t,data);
  await card(page,0).locator('select').selectOption('disabled');await page.locator('#plugin-next').click();await card(page,20).locator('select').selectOption('enabled');
  await page.locator('#plugin-only-drafts').check();await page.locator('#plugin-search').fill('/unique-0');assert.deepEqual(await names(page),['plugin-000']);
  let message='';const capture=dialog=>{message=dialog.message();};page.on('dialog',capture);
  await confirmAction(page,false,()=>page.locator('#discard-plugin-drafts').click());assert.match(message,/2 个插件/);assert.equal(await page.evaluate(()=>pluginDrafts.size),2);
  assert.equal(await card(page,0).locator('select').inputValue(),'disabled');assert.equal(await page.evaluate(()=>requests.length),0);
  await confirmAction(page,true,()=>page.locator('#discard-plugin-drafts').click());page.off('dialog',capture);
  assert.equal(await page.evaluate(()=>pluginDrafts.size),0);assert.equal(await page.locator('#discard-plugin-drafts').isDisabled(),true);assert.match(await page.locator('#plugin-save-state').textContent(),/已放弃 2 个.*已保存的开关规则未改变/);
  assert.match(await page.locator('#plugin-empty').textContent(),/当前范围没有未保存/);
  await page.locator('#clear-plugin-filter').click();assert.equal(await card(page,0).locator('select').inputValue(),'inherit');assert.equal(await card(page,1).locator('select').inputValue(),'disabled');
  assert.equal(await page.evaluate(()=>plugins[1].state.effectiveEnabled),false);await page.locator('#plugin-next').click();assert.equal(await card(page,20).locator('select').inputValue(),'inherit');
  assert.equal(await page.evaluate(()=>requests.length),0);
});

test('saving freezes pending actions, clears only the saved draft, clamps pages, and retains failed drafts and revisions',async t=>{
  const page=await harness(t,fixtures(21));
  await page.evaluate(()=>{for(const p of plugins)pluginDrafts.set(p.name,{mode:'disabled',revision:p.state.revision});renderPlugins();});
  await page.locator('#plugin-only-drafts').check();await page.locator('#plugin-next').click();await page.evaluate(()=>replies.push({pending:true}));
  await card(page,20).locator('.plugin-controls button').click();await page.waitForFunction(()=>pending.length===1);
  assert.equal(await page.locator('#plugin-only-drafts').isDisabled(),true);assert.equal(await page.locator('#discard-plugin-drafts').isDisabled(),true);
  await page.evaluate(()=>discardPluginDrafts());assert.equal(await page.evaluate(()=>pluginDrafts.size),21);
  await page.evaluate(saved=>pending[0]({ok:true,status:200,json:async()=>({state:saved})}),state(false,'disabled'));await idle(page);
  assert.equal(await page.evaluate(()=>pluginDrafts.size),20);assert.equal(await page.evaluate(()=>pluginPage),1);assert.equal(await card(page,20).count(),0);assert.equal(await page.locator('#plugin-only-drafts').isDisabled(),false);
  await page.evaluate(()=>replies.push({error:'版本冲突',status:409}));await card(page,0).locator('.plugin-controls button').click();await idle(page);
  assert.equal(await page.evaluate(()=>pluginDrafts.get('plugin-000').revision),'a'.repeat(64));assert.equal(await card(page,0).locator('select').inputValue(),'disabled');assert.equal(await page.evaluate(()=>pluginDrafts.size),20);
  await confirmAction(page,true,()=>page.locator('#discard-plugin-drafts').click());await page.locator('#clear-plugin-filter').click();await page.locator('#plugin-next').click();
  assert.equal(await card(page,20).locator('select').inputValue(),'disabled');assert.equal(await page.evaluate(()=>plugins[20].state.effectiveEnabled),false);assert.equal(await page.evaluate(()=>requests.length),2);
});

test('scope read failures preserve pending criteria, successful reads clear drafts, and old responses cannot revive expired-session choices',async t=>{
  const page=await harness(t,fixtures(21));await card(page,0).locator('select').selectOption('disabled');await page.locator('#plugin-only-drafts').check();
  await page.selectOption('#plugin-scope','group');await page.locator('#plugin-group').fill('234567890');await page.evaluate(()=>replies.push({error:'读取失败',status:500}));
  await confirmAction(page,true,()=>page.locator('#apply-plugin-scope').click());await idle(page);assert.equal(await page.evaluate(()=>pluginGroup),'0');assert.equal(await page.locator('#plugin-only-drafts').isChecked(),true);assert.equal(await page.evaluate(()=>pluginDrafts.size),1);
  await page.selectOption('#plugin-scope','group');await page.locator('#plugin-group').fill('234567890');await page.evaluate(data=>replies.push({plugins:data,controlAvailable:true}),fixtures(21));
  await confirmAction(page,true,()=>page.locator('#apply-plugin-scope').click());await idle(page);assert.equal(await page.evaluate(()=>pluginGroup),'234567890');assert.equal(await page.evaluate(()=>pluginDrafts.size),0);assert.equal(await page.locator('#plugin-only-drafts').isChecked(),true);assert.match(await page.locator('#plugin-empty').textContent(),/当前范围没有未保存/);
  await page.locator('#clear-plugin-filter').click();await card(page,0).locator('select').selectOption('disabled');await page.locator('#plugin-only-drafts').check();await page.evaluate(()=>replies.push({pending:true}));
  await confirmAction(page,true,()=>page.locator('#reload-plugins').click());await page.waitForFunction(()=>pending.length===1);
  await page.evaluate(data=>{authenticated=false;csrf='';showAuth();authenticated=true;csrf='new-session';plugins=data;pluginControlAvailable=true;document.getElementById('auth').hidden=true;document.getElementById('workspace').hidden=false;renderPlugins();},fixtures(21));
  assert.equal(await page.locator('#plugin-only-drafts').isChecked(),false);assert.equal(await page.locator('#discard-plugin-drafts').isDisabled(),true);
  await card(page,1).locator('select').selectOption('disabled');await page.locator('#plugin-only-drafts').check();
  await page.evaluate(data=>pending[0]({ok:true,status:200,json:async()=>({plugins:data,controlAvailable:true})}),fixtures(1));await page.waitForTimeout(30);
  assert.equal(await page.locator('#plugin-only-drafts').isChecked(),true);assert.deepEqual(await names(page),['plugin-001']);assert.equal(await page.evaluate(()=>pluginDrafts.size),1);
  await page.evaluate(()=>replies.push({error:'过期',status:401}));await card(page,1).locator('.plugin-controls button').click();await page.waitForFunction(()=>!authenticated);
  assert.equal(await page.locator('#plugin-only-drafts').isChecked(),false);assert.equal(await page.evaluate(()=>pluginOnlyDrafts),false);assert.equal(await page.evaluate(()=>pluginDrafts.size),0);
  await page.evaluate(data=>{authenticated=true;csrf='last-session';plugins=data;pluginControlAvailable=true;document.getElementById('auth').hidden=true;document.getElementById('workspace').hidden=false;renderPlugins();replies.push({});},fixtures(21));
  await card(page,0).locator('select').selectOption('disabled');await page.locator('#plugin-only-drafts').check();await confirmAction(page,true,()=>page.locator('#logout').click());await page.waitForFunction(()=>!authenticated);
  assert.equal(await page.locator('#plugin-only-drafts').isChecked(),false);assert.equal(await page.evaluate(()=>pluginDrafts.size),0);
});
