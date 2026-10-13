// Real Chrome DOM with synthetic data only; all network access is blocked.
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
const state=(enabled,globalMode='inherit',groupMode='inherit',source='default')=>({effectiveEnabled:enabled,defaultEnabled:true,globalMode,groupMode,source,revision:'a'.repeat(64)});
const fixtures=[
  {name:'llm',brief:'GPT 与群总结',usage:'/gpt 你好\n总结最近消息',state:state(true)},
  {name:'sleepmanage',brief:'睡眠小助手',usage:'早安\n晚安',state:state(false,'disabled','enabled','global')},
  {name:'Legacy',brief:'旧插件',usage:'<img src=x onerror=bad()>\nHELLO😀',defaultEnabledAtStartup:true}
];
async function harness(t){
  const context=await browser.newContext(); t.after(()=>context.close());
  await context.route('**/*',route=>route.abort());
  const page=await context.newPage(); await page.setContent(html); await page.addStyleTag({content:css});
  await page.addScriptTag({content:app});
  await page.evaluate(data=>{
    authenticated=true; csrf='fixture-session'; plugins=data; pluginControlAvailable=true;
    window.showOverview=()=>{}; window.showLogs=()=>{}; window.showBackups=()=>{};
    document.getElementById('auth').hidden=true; document.getElementById('workspace').hidden=false;
    document.getElementById('overview-view').hidden=true; document.getElementById('plugins-view').hidden=false;
    window.apiRequests=[]; window.apiReplies=[]; window.pending=[];
    window.fetch=async(path,options)=>{
      window.apiRequests.push({path,body:options.body?JSON.parse(options.body):null});
      if (!window.apiReplies.length) throw new Error('Missing API fixture');
      const body=window.apiReplies.shift();
      if (body.pending) return new Promise(resolve=>pending.push(resolve));
      return {ok:!body.error,status:body.status||200,json:async()=>body};
    };
    renderPlugins();
  },fixtures);
  return page;
}
const names=page=>page.locator('#plugin-list h2').allTextContents();

test('searches name, brief and usage with combined actual-state filter and full counts',async t=>{
  const page=await harness(t);
  assert.deepEqual(await names(page),['llm','sleepmanage','Legacy']);
  assert.equal(await page.locator('#search-count').textContent(),'3 / 3 个插件；启用 1，禁用 1，未知 1');
  await page.locator('#plugin-search').fill('  /GPT  '); assert.deepEqual(await names(page),['llm']);
  await page.selectOption('#plugin-state-filter','disabled'); assert.deepEqual(await names(page),[]);
  assert.equal(await page.locator('#plugin-empty').isVisible(),true);
  assert.match(await page.locator('#plugin-empty').textContent(),/清除筛选/);
  await page.locator('#clear-plugin-filter').click();
  await page.locator('#plugin-search').fill('睡眠'); assert.deepEqual(await names(page),['sleepmanage']);
  await page.locator('#plugin-search').fill('早安'); assert.deepEqual(await names(page),['sleepmanage']);
  await page.locator('#clear-plugin-filter').click();
  await page.selectOption('#plugin-state-filter','unknown'); assert.deepEqual(await names(page),['Legacy']);
  await page.locator('#plugin-search').fill('hello😀'); assert.deepEqual(await names(page),['Legacy']);
  await page.locator('#plugin-list .text-button').click();
  assert.equal(await page.locator('#usage-text').textContent(),fixtures[2].usage);
  assert.equal(await page.locator('#usage-text img').count(),0);
  assert.equal(await page.evaluate(()=>apiRequests.length),0);
});

test('filtering preserves drafts, uses committed state and re-filters after a saved switch',async t=>{
  const page=await harness(t);
  await page.selectOption('#plugin-state-filter','enabled');
  await page.locator('#plugin-list .plugin-controls select').selectOption('disabled');
  assert.deepEqual(await names(page),['llm']); // Draft cannot count as effective state.
  await page.selectOption('#plugin-state-filter','disabled'); assert.deepEqual(await names(page),['sleepmanage']);
  await page.locator('#clear-plugin-filter').click();
  assert.equal(await page.locator('.plugin-card').filter({has:page.locator('h2',{hasText:'llm'})}).locator('select').inputValue(),'disabled');
  assert.equal(await page.evaluate(()=>apiRequests.length),0);
  await page.selectOption('#plugin-state-filter','enabled');
  await page.evaluate(saved=>apiReplies.push({state:saved}),state(false,'disabled','inherit','global'));
  await page.locator('#plugin-list .plugin-controls button').click();
  await page.waitForFunction(()=>!pluginSaving);
  assert.deepEqual(await names(page),[]);
  assert.match(await page.locator('#search-count').textContent(),/启用 0，禁用 2，未知 1/);
  const requests=await page.evaluate(()=>apiRequests);
  assert.equal(requests.length,1); assert.equal(requests[0].path,'/api/plugin-switch');
  assert.deepEqual(requests[0].body,{name:'llm',groupId:'0',mode:'disabled',revision:'a'.repeat(64)});
  await page.locator('#clear-plugin-filter').click();
  assert.equal(await page.evaluate(()=>pluginDrafts.size),0);
});

test('group scope updates actual state while retaining search, filters and inherited precedence',async t=>{
  const page=await harness(t);
  await page.locator('#plugin-search').fill('早安');
  await page.selectOption('#plugin-state-filter','enabled'); assert.deepEqual(await names(page),[]);
  await page.selectOption('#plugin-scope','group'); await page.locator('#plugin-group').fill('123456789');
  const data=structuredClone(fixtures); data[1].state=state(true,'inherit','enabled','group');
  await page.evaluate(data=>apiReplies.push({plugins:data,groupId:'123456789',controlAvailable:true}),data);
  await page.locator('#apply-plugin-scope').click(); await page.waitForFunction(()=>!pluginScopeBusy);
  assert.deepEqual(await names(page),['sleepmanage']);
  assert.equal(await page.locator('#plugin-state-filter').inputValue(),'enabled');
  assert.equal(await page.locator('#plugin-search').inputValue(),'早安');
  assert.match(await page.locator('#plugin-list .field-help:not(.plugin-draft-note)').textContent(),/此群设置/);
  assert.match(await page.locator('#plugin-scope-state').textContent(),/123456789/);
  assert.equal((await page.evaluate(()=>apiRequests))[0].path,'/api/plugins?group=123456789');
});

test('distinguishes empty registry and unknown state and clears criteria when logging out',async t=>{
  const page=await harness(t);
  await page.evaluate(()=>{plugins=plugins.map(p=>({...p,state:undefined})); renderPlugins();});
  await page.selectOption('#plugin-state-filter','enabled'); assert.deepEqual(await names(page),[]);
  await page.selectOption('#plugin-state-filter','unknown'); assert.equal((await names(page)).length,3);
  assert.match(await page.locator('#search-count').textContent(),/启用 0，禁用 0，未知 3/);
  await page.evaluate(()=>{plugins=[]; renderPlugins();});
  assert.equal(await page.locator('#plugin-empty').textContent(),'当前程序未注册插件。');
  await page.locator('#plugin-search').fill('private query');
  await page.evaluate(()=>{authenticated=false;showAuth();});
  assert.equal(await page.locator('#plugin-search').inputValue(),'');
  assert.equal(await page.locator('#plugin-state-filter').inputValue(),'all');
  assert.equal(await page.locator('#auth').isVisible(),true);
});

const card=(page,name)=>page.locator('.plugin-card').filter({has:page.locator('h2',{hasText:name})});
async function confirmAction(page,accept,action){
  const dialog=page.waitForEvent('dialog'); const result=action();
  const d=await dialog; assert.equal(d.type(),'confirm'); assert.match(d.message(),/未保存/);
  if(accept) await d.accept(); else await d.dismiss(); await result;
}
const idle=page=>page.waitForFunction(()=>!pluginSaving&&!pluginScopeBusy);

test('marks only changes, counts hidden drafts and preserves drafts across page navigation',async t=>{
  const page=await harness(t);
  await card(page,'llm').locator('select').selectOption('disabled');
  await card(page,'sleepmanage').locator('select').selectOption('enabled');
  assert.match(await page.locator('#plugin-draft-status').textContent(),/2 个插件（全局）/);
  assert.equal(await page.locator('.plugin-draft-note:visible').count(),2);
  await page.locator('#plugin-search').fill('/gpt');
  assert.match(await page.locator('#plugin-draft-status').textContent(),/2 个插件/);
  await page.locator('[data-view="overview"]').click(); await page.locator('[data-view="plugins"]').click();
  assert.equal(await card(page,'llm').locator('select').inputValue(),'disabled');
  await card(page,'llm').locator('select').selectOption('inherit');
  assert.match(await page.locator('#plugin-draft-status').textContent(),/1 个插件/);
  await page.locator('#clear-plugin-filter').click();
  await card(page,'sleepmanage').locator('select').selectOption('disabled');
  assert.equal(await page.evaluate(()=>pluginDrafts.size),0);
  assert.equal(await page.evaluate(()=>apiRequests.length),0);
});

test('cancelled reload, scope and logout send nothing; failed read preserves scope, mode and revision',async t=>{
  const page=await harness(t); await card(page,'llm').locator('select').selectOption('disabled');
  await confirmAction(page,false,()=>page.locator('#reload-plugins').click());
  await page.selectOption('#plugin-scope','group'); await page.locator('#plugin-group').fill('234567890');
  await confirmAction(page,false,()=>page.locator('#apply-plugin-scope').click());
  await confirmAction(page,false,()=>page.locator('#logout').click());
  assert.equal(await page.evaluate(()=>apiRequests.length),0);
  await page.selectOption('#plugin-scope','group'); await page.locator('#plugin-group').fill('234567890');
  await page.evaluate(()=>apiReplies.push({error:'模拟读取失败',status:500}));
  await confirmAction(page,true,()=>page.locator('#apply-plugin-scope').click()); await idle(page);
  assert.equal(await page.evaluate(()=>pluginGroup),'0');
  assert.equal(await page.locator('#plugin-scope').inputValue(),'global');
  assert.equal(await card(page,'llm').locator('select').inputValue(),'disabled');
  assert.equal(await page.evaluate(()=>pluginDrafts.get('llm').revision),'a'.repeat(64));
  const next=structuredClone(fixtures); next[0].state=state(false,'inherit','disabled','group');
  await page.selectOption('#plugin-scope','group'); await page.locator('#plugin-group').fill('234567890');
  await page.evaluate(data=>apiReplies.push({plugins:data,controlAvailable:true}),next);
  await confirmAction(page,true,()=>page.locator('#apply-plugin-scope').click()); await idle(page);
  assert.equal(await page.evaluate(()=>pluginDrafts.size),0);
  assert.equal(await page.evaluate(()=>pluginGroup),'234567890');
  assert.equal(await card(page,'llm').locator('select').inputValue(),'disabled');
});

test('individual save clears only its draft; conflict preserves remaining draft and confirmed reload clears it',async t=>{
  const page=await harness(t);
  await card(page,'llm').locator('select').selectOption('disabled');
  await card(page,'sleepmanage').locator('select').selectOption('enabled');
  await page.evaluate(saved=>apiReplies.push({state:saved}),state(false,'disabled','inherit','global'));
  await card(page,'llm').locator('.plugin-controls button').click(); await idle(page);
  assert.deepEqual(await page.evaluate(()=>[...pluginDrafts.keys()]),['sleepmanage']);
  await page.evaluate(()=>apiReplies.push({error:'配置版本冲突',status:409}));
  await card(page,'sleepmanage').locator('.plugin-controls button').click(); await idle(page);
  assert.equal(await card(page,'sleepmanage').locator('select').inputValue(),'enabled');
  assert.match(await page.locator('#plugin-error').textContent(),/冲突/);
  await page.evaluate(data=>apiReplies.push({plugins:data,controlAvailable:true}),fixtures);
  await confirmAction(page,true,()=>page.locator('#reload-plugins').click()); await idle(page);
  assert.equal(await page.evaluate(()=>pluginDrafts.size),0);
});

test('busy requests prevent logout and browser departure prompts; successful logout and expiry reset scope and drafts',async t=>{
  const page=await harness(t);
  const unload=()=>page.evaluate(()=>{const event=new Event('beforeunload',{cancelable:true});window.dispatchEvent(event);return event.defaultPrevented;});
  assert.equal(await unload(),false);
  await card(page,'llm').locator('select').selectOption('disabled'); assert.equal(await unload(),true);
  await page.evaluate(()=>apiReplies.push({pending:true}));
  await card(page,'llm').locator('.plugin-controls button').click(); await page.waitForFunction(()=>pending.length===1);
  await page.locator('#logout').click(); assert.equal(await page.evaluate(()=>apiRequests.length),1);
  assert.equal(await page.locator('#reload-plugins').isDisabled(),true);
  await page.evaluate(saved=>pending[0]({ok:true,status:200,json:async()=>({state:saved})}),state(false,'disabled','inherit','global'));await idle(page);
  assert.equal(await unload(),false);
  await card(page,'llm').locator('select').selectOption('inherit');
  await page.evaluate(()=>apiReplies.push({})); await confirmAction(page,true,()=>page.locator('#logout').click());
  await page.waitForFunction(()=>!authenticated);
  assert.equal(await page.evaluate(()=>pluginDrafts.size),0); assert.equal(await page.locator('#plugin-list').textContent(),'');
  assert.equal(await unload(),false);
  await page.evaluate(data=>{authenticated=true;csrf='expiry-session';plugins=data;pluginControlAvailable=true;pluginGroup='123456789';renderPluginScope();renderPluginBusy();renderPlugins();document.getElementById('auth').hidden=true;document.getElementById('workspace').hidden=false;},fixtures);
  await card(page,'llm').locator('select').selectOption('disabled');
  await page.evaluate(()=>apiReplies.push({error:'会话过期',status:401}));
  await card(page,'llm').locator('.plugin-controls button').click(); await page.waitForFunction(()=>!authenticated);
  assert.equal(await page.evaluate(()=>pluginGroup),'0'); assert.equal(await page.locator('#plugin-group').inputValue(),'');
  assert.equal(await page.evaluate(()=>pluginDrafts.size),0); assert.equal(await page.locator('#plugin-error').isVisible(),false);
});

test('late save and old 401 cannot change, clear or unlock a new session scope request',async t=>{
  const page=await harness(t);
  await card(page,'llm').locator('select').selectOption('disabled');
  await page.evaluate(()=>apiReplies.push({pending:true}));
  await card(page,'llm').locator('.plugin-controls button').click(); await page.waitForFunction(()=>pending.length===1);
  await page.evaluate(data=>{authenticated=false;csrf='';showAuth();authenticated=true;csrf='new-session';plugins=data;pluginControlAvailable=true;document.getElementById('auth').hidden=true;document.getElementById('workspace').hidden=false;renderPlugins();apiReplies.push({pending:true});},fixtures);
  await page.locator('#reload-plugins').click(); await page.waitForFunction(()=>pending.length===2);
  await page.evaluate(saved=>pending[0]({ok:true,status:200,json:async()=>({state:saved})}),state(false,'disabled','inherit','global'));
  await page.waitForTimeout(30);
  assert.equal(await page.evaluate(()=>pluginScopeBusy),true); assert.equal(await page.locator('#reload-plugins').isDisabled(),true);
  assert.equal(await page.evaluate(()=>plugins[0].state.effectiveEnabled),true);
  await page.evaluate(()=>{authenticated=false;csrf='';showAuth();authenticated=true;csrf='third-session';document.getElementById('auth').hidden=true;document.getElementById('workspace').hidden=false;apiReplies.push({pending:true});});
  await page.locator('#reload-plugins').click(); await page.waitForFunction(()=>pending.length===3);
  await page.evaluate(()=>pending[1]({ok:false,status:401,json:async()=>({error:'旧会话过期'})}));await page.waitForTimeout(30);
  assert.equal(await page.evaluate(()=>authenticated),true); assert.equal(await page.evaluate(()=>pluginScopeBusy),true);
  await page.evaluate(data=>pending[2]({ok:true,status:200,json:async()=>({plugins:data,controlAvailable:true})}),fixtures);await idle(page);
  assert.equal(await page.locator('#plugin-error').isVisible(),false);
  assert.equal(await page.evaluate(()=>plugins.length),3);
});
