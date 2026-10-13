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

const discard=(p,i)=>card(p,i).locator('.plugin-discard-one');
async function confirmOne(p,i,accept=true){
 const event=p.waitForEvent('dialog'),click=discard(p,i).click();const dialog=await event;const message=dialog.message();
 assert.equal(dialog.type(),'confirm');if(accept)await dialog.accept();else await dialog.dismiss();await click;return message;
}

test('cancel and confirmation target only the loaded plugin/scope, keeping hidden drafts and actual saved rules',async t=>{
 const data=fixtures(41);data[0].state=state(false,'disabled');const p=await harness(t,data);
 assert.equal(await discard(p,0).isDisabled(),true);await card(p,0).locator('select').selectOption('enabled');
 await p.locator('#plugin-next').click();await card(p,20).locator('select').selectOption('disabled');await p.locator('#plugin-next').click();await card(p,40).locator('select').selectOption('enabled');
 await p.locator('#plugin-prev').click();await p.locator('#plugin-prev').click();await p.selectOption('#plugin-scope','group');await p.locator('#plugin-group').fill('987654321');
 const before=await p.evaluate(()=>JSON.stringify(plugins));let message=await confirmOne(p,0,false);assert.match(message,/plugin-000.*全局.*其他插件草稿保留/);assert.doesNotMatch(message,/987654321/);assert.equal(await card(p,0).locator('select').inputValue(),'enabled');assert.equal(await p.evaluate(()=>pluginDrafts.size),3);
 await confirmOne(p,0);assert.equal(await card(p,0).locator('select').inputValue(),'disabled');assert.equal(await discard(p,0).isDisabled(),true);assert.deepEqual(await p.evaluate(()=>[...pluginDrafts.keys()]),['plugin-020','plugin-040']);assert.equal(await p.evaluate(()=>JSON.stringify(plugins)),before);assert.equal(await p.evaluate(()=>requests.length),0);assert.match(await p.locator('#plugin-save-state').textContent(),/已放弃「plugin-000」.*全局/);
});

test('only-draft pagination shrinks and draft-priority order reflows while filters and unrelated drafts persist',async t=>{
 const p=await harness(t,fixtures(21));await p.evaluate(()=>{for(let i=0;i<11;i++)pluginDrafts.set(plugins[i].name,{mode:'disabled',revision:plugins[i].state.revision});renderPlugins()});
 await p.locator('#plugin-only-drafts').check();await p.selectOption('#plugin-page-size','10');await p.selectOption('#plugin-sort','drafts');await p.locator('#plugin-next').click();assert.deepEqual(await names(p),['plugin-010']);await confirmOne(p,10);
 assert.equal(await p.evaluate(()=>pluginPage),1);assert.equal((await names(p)).length,10);assert.equal(await p.locator('#plugin-only-drafts').isChecked(),true);assert.equal(await p.locator('#plugin-sort').inputValue(),'drafts');assert.equal(await p.locator('#plugin-page-size').inputValue(),'10');assert.equal(await p.evaluate(()=>pluginDrafts.size),10);
 await p.locator('#plugin-only-drafts').uncheck();await p.locator('#plugin-search').fill('unique-0');await confirmOne(p,0);assert.equal(await discard(p,0).isDisabled(),true);assert.equal(await p.locator('#plugin-search').inputValue(),'unique-0');await p.locator('#clear-plugin-filter').click();assert.equal((await names(p))[0],'plugin-001');assert.equal(await p.evaluate(()=>pluginDrafts.size),9);assert.equal(await p.evaluate(()=>requests.length),0);
});

test('discard uses the last successful save or read baseline and does not clear conflict errors or undo saved changes',async t=>{
 const p=await harness(t,fixtures(3));await card(p,0).locator('select').selectOption('disabled');const saved={...state(false,'disabled'),source:'global',revision:'b'.repeat(64)};
 await p.evaluate(state=>replies.push({state}),saved);await card(p,0).locator('.plugin-controls button').click();await idle(p);await card(p,0).locator('select').selectOption('enabled');await confirmOne(p,0);assert.equal(await card(p,0).locator('select').inputValue(),'disabled');assert.deepEqual(await p.evaluate(()=>plugins[0].state),saved);assert.equal(await p.evaluate(()=>requests.length),1);
 await card(p,1).locator('select').selectOption('disabled');await p.evaluate(()=>replies.push({error:'fixture version conflict',status:409}));await card(p,1).locator('.plugin-controls button').click();await idle(p);await confirmOne(p,1);assert.equal(await card(p,1).locator('select').inputValue(),'inherit');assert.match(await p.locator('#plugin-error').textContent(),/version conflict/);assert.equal(await p.evaluate(()=>plugins[1].state.revision),'a'.repeat(64));assert.equal(await p.evaluate(()=>requests.length),2);
 await p.evaluate(()=>{pluginGroup='12345678';plugins[2].state.groupMode='disabled';pluginDrafts.set(plugins[2].name,{mode:'enabled',revision:plugins[2].state.revision});renderPlugins()});assert.match(await confirmOne(p,2),/群 12345678/);assert.equal(await card(p,2).locator('select').inputValue(),'disabled');
});

test('no draft/unavailable/busy controls cannot discard; confirmation becomes invalid after target, draft, baseline or session changes',async t=>{
 const p=await harness(t,fixtures(3,true));assert.equal(await discard(p,0).isDisabled(),true);assert.equal(await card(p,2).locator('.plugin-discard-one').count(),0);await card(p,0).locator('select').selectOption('disabled');
 for(const busy of ['pluginSaving','pluginScopeBusy']){await p.evaluate(busy=>{window.eval(busy+'=true');renderPluginBusy()},busy);assert.equal(await discard(p,0).isDisabled(),true);await p.evaluate(()=>discardOnePluginDraft(plugins[0]));assert.equal(await p.evaluate(()=>pluginDrafts.size),1);await p.evaluate(busy=>{window.eval(busy+'=false');renderPluginBusy()},busy);}
 for(const mutation of ['draft','baseline','list','group','token','generation','busy','logout']){
  await p.evaluate(mutation=>{
   authenticated=true;csrf='fixture-session';pluginGroup='0';pluginSaving=false;pluginScopeBusy=false;plugins[0].state.revision='a'.repeat(64);pluginDrafts.set(plugins[0].name,{mode:'disabled',revision:'a'.repeat(64)});document.getElementById('plugin-save-state').textContent='unchanged';
   window.confirm=()=>{switch(mutation){case 'draft':pluginDrafts.get(plugins[0].name).mode='enabled';break;case 'baseline':plugins[0].state.revision='b'.repeat(64);break;case 'list':plugins=[...plugins];break;case 'group':pluginGroup='999';break;case 'token':csrf='new-token';break;case 'generation':pluginGeneration++;break;case 'busy':pluginSaving=true;break;case 'logout':authenticated=false;break;}return true};discardOnePluginDraft(plugins[0]);
  },mutation);assert.equal(await p.evaluate(()=>pluginDrafts.size),1);assert.equal(await p.locator('#plugin-save-state').textContent(),'unchanged');
 }
 await p.evaluate(()=>{authenticated=true;pluginSaving=false;pluginScopeBusy=false;window.confirm=()=>{pluginDrafts.set('renamed-plugin',{...pluginDrafts.get(plugins[0].name)});plugins[0].name='renamed-plugin';return true};discardOnePluginDraft(plugins[0])});assert.equal(await p.evaluate(()=>pluginDrafts.size),2);
 assert.equal(await p.evaluate(()=>requests.length),0);
});

test('expiry removes actions and stale reads cannot overwrite a new session or discard its new same-name draft',async t=>{
 for(const status of [200,401]){
  const p=await harness(t,fixtures(3));await p.evaluate(()=>replies.push({pending:true}));await p.locator('#reload-plugins').click();await p.waitForFunction(()=>pending.length===1);
  await p.evaluate(data=>{window.oldPlugin=plugins[0];authenticated=false;csrf='';showAuth();authenticated=true;csrf='new-session';plugins=data;pluginControlAvailable=true;document.getElementById('auth').hidden=true;document.getElementById('workspace').hidden=false;renderPluginBusy();renderPlugins()},fixtures(3));await card(p,0).locator('select').selectOption('disabled');
  await p.evaluate(({status,data})=>pending[0]({ok:status===200,status,json:async()=>status===200?{plugins:data,controlAvailable:true}:{error:'old expiry'}}),{status,data:fixtures(21)});await p.waitForTimeout(40);await p.evaluate(()=>discardOnePluginDraft(oldPlugin));assert.equal(await p.evaluate(()=>authenticated),true);assert.equal(await p.evaluate(()=>pluginDrafts.size),1);assert.equal(await discard(p,0).isDisabled(),false);
  await p.evaluate(()=>replies.push({error:'current expiry',status:401}));await confirmAction(p,true,()=>p.locator('#reload-plugins').click());await p.waitForFunction(()=>!authenticated);assert.equal(await p.locator('.plugin-discard-one').count(),0);assert.equal(await p.evaluate(()=>pluginDrafts.size),0);
 }
});

