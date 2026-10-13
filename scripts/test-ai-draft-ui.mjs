// Offline Chrome tests: synthetic settings, mocked APIs, no actual secrets or QQ.
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
const ai=await fs.readFile('internal/localadmin/assets/ai.js','utf8');
const browser=await chromium.launch({channel:'chrome',headless:true});
after(()=>browser.close());
const group=(service,temperature=70)=>({service,groupId:'123456789',revision:'b'.repeat(64),temperature,rate:0,useAgent:true,voiceOutput:false,respondAt:true});
function fixture(){return {available:true,global:{fields:{Type:0,ModelName:'fixture',API:'http://fixture.invalid/chat',MaxN:4096,TopP:0.9,SystemP:'fixture prompt',NoSystemP:false,ReasoningEffort:'none',Separator:'',ImageType:0,ImageModelName:'',ImageAPI:'',AgentType:0,AgentModelName:'fixture',AgentAPI:'http://fixture.invalid/agent',AgentChar:'可爱、略带傲娇',AgentSex:'女'},keys:{Key:true,ImageKey:false,AgentKey:true},addressRedacted:{},revision:'a'.repeat(64),ready:true},groups:[group('aichatcfg'),group('llm')]};}
async function harness(t){
  const context=await browser.newContext(); t.after(()=>context.close());
  await context.route('**/*',route=>route.abort()); const page=await context.newPage();await page.setContent(html);
  await page.addScriptTag({content:app}); await page.addScriptTag({content:ai});
  await page.evaluate(data=>{
    authenticated=true;csrf='fixture-session';aiState=data;aiGroup='123456789';
    window.showOverview=()=>{};window.showLogs=()=>{};window.showBackups=()=>{};
    document.getElementById('auth').hidden=true;document.getElementById('workspace').hidden=false;
    for(const section of document.querySelectorAll('.content>section')) section.hidden=section.id!=='ai-view';
    for(const button of document.querySelectorAll('.nav-button')) button.classList.toggle('active',button.dataset.view==='ai');
    window.requests=[];window.replies=[];window.pending=[];
    window.fetch=async(url,options)=>{
      requests.push({url,body:options.body?JSON.parse(options.body):null});
      const reply=replies.shift();
      if (!reply) return new Promise(resolve=>pending.push(resolve));
      return {ok:!reply.error,status:reply.status||200,json:async()=>reply};
    };
    renderAIGlobal(data);renderAIGroups(data.groups);aiBusyState(false);
  },fixture());return page;
}
const dirty=page=>page.evaluate(()=>aiDrafts());
const waitIdle=page=>page.waitForFunction(()=>!aiBusy);
async function dialogAction(page,accept,action){
  const dialog=page.waitForEvent('dialog');const result=action();
  const d=await dialog;assert.equal(d.type(),'confirm');if(accept)await d.accept();else await d.dismiss();await result;
}

test('tracks reverted fields, preserves navigation drafts and keeps key values out of status and storage',async t=>{
  const page=await harness(t);assert.deepEqual(await dirty(page),[]);
  await page.locator('#ai-field-AgentChar').fill('新的设定😀');assert.deepEqual(await dirty(page),['global']);
  await page.locator('#ai-field-AgentChar').fill(fixture().global.fields.AgentChar);assert.deepEqual(await dirty(page),[]);
  await page.selectOption('#ai-action-Key','replace');await page.locator('#ai-key-Key').fill('synthetic-private-key');
  assert.deepEqual(await dirty(page),['global']);assert.equal(await page.locator('#ai-key-Key').getAttribute('type'),'password');
  assert.doesNotMatch(await page.locator('#ai-draft-status').textContent(),/synthetic-private-key/);
  await page.locator('[data-view="plugins"]').click();await page.locator('[data-view="ai"]').click();
  assert.equal(await page.locator('#ai-key-Key').inputValue(),'synthetic-private-key');
  assert.equal(await page.evaluate(()=>requests.length),0);
  // Opaque offline pages have no persistent storage access; inspect the shipped controller too.
  assert.doesNotMatch(ai,/localStorage|sessionStorage|indexedDB/);
  await page.selectOption('#ai-action-Key','keep');assert.deepEqual(await dirty(page),[]);
});

test('cancelled reload/group switch preserves edits and failed reads keep the old draft',async t=>{
  const page=await harness(t);await page.locator('#ai-field-AgentChar').fill('保留这段文字');
  await dialogAction(page,false,()=>page.locator('#reload-ai').click());assert.equal(await page.evaluate(()=>requests.length),0);
  await page.locator('#ai-group').fill('234567890');
  await dialogAction(page,false,()=>page.locator('#apply-ai-group').click());assert.equal(await page.evaluate(()=>aiGroup),'123456789');
  await page.evaluate(()=>replies.push({error:'模拟读取故障',status:500}));
  await dialogAction(page,true,()=>page.locator('#reload-ai').click());await waitIdle(page);
  assert.equal(await page.locator('#ai-field-AgentChar').inputValue(),'保留这段文字');assert.deepEqual(await dirty(page),['global']);
  const next=fixture();next.groups=next.groups.map(g=>({...g,groupId:'234567890'}));
  await page.evaluate(next=>replies.push(next),next);
  await dialogAction(page,true,()=>page.locator('#apply-ai-group').click());await waitIdle(page);
  assert.equal(await page.evaluate(()=>aiGroup),'234567890');assert.deepEqual(await dirty(page),[]);
});

test('successful saves clear only the saved form and conflict/validation failures retain drafts',async t=>{
  const page=await harness(t);
  await page.locator('#ai-field-AgentChar').fill('全局草稿');
  await page.locator('#ai-field-temperature-aichatcfg').fill('80');await page.locator('#ai-field-temperature-llm').fill('90');
  assert.deepEqual(await dirty(page),['global','aichatcfg','llm']);
  const saved=fixture();saved.global.fields.AgentChar='全局草稿';saved.global.revision='c'.repeat(64);
  await page.evaluate(saved=>replies.push(saved),saved);await page.locator('#save-ai').click();await waitIdle(page);
  assert.deepEqual(await dirty(page),['aichatcfg','llm']);assert.equal(await page.locator('#ai-field-temperature-llm').inputValue(),'90');
  const groups=[group('aichatcfg',80),group('llm')];
  await page.evaluate(groups=>replies.push({groups}),groups);
  await page.locator('#ai-group-forms form[data-service="aichatcfg"] button[type="submit"]').click();await waitIdle(page);
  assert.deepEqual(await dirty(page),['llm']);assert.equal(await page.locator('#ai-field-temperature-llm').inputValue(),'90');
  await page.evaluate(()=>replies.push({error:'配置版本已变化',status:409}));
  await page.locator('#ai-group-forms form[data-service="llm"] button[type="submit"]').click();await waitIdle(page);
  assert.deepEqual(await dirty(page),['llm']);assert.match(await page.locator('#ai-error').textContent(),/版本/);
  await page.selectOption('#ai-action-Key','clear');await page.locator('#save-ai').click();
  assert.deepEqual(await dirty(page),['global','llm']);assert.equal(await page.evaluate(()=>requests.length),3);
  assert.match(await page.locator('#ai-error').textContent(),/确认清除/);
});

test('logout/expiry clears drafts and stale responses cannot unlock or expire a new session',async t=>{
  const page=await harness(t);await page.locator('#ai-field-AgentChar').fill('旧草稿');
  await dialogAction(page,false,()=>page.locator('#logout').click());assert.equal(await page.evaluate(()=>requests.length),0);
  await page.evaluate(()=>replies.push({}));
  await dialogAction(page,true,()=>page.locator('#logout').click());
  await page.waitForFunction(()=>!authenticated);assert.deepEqual(await dirty(page),[]);
  assert.equal(await page.locator('#ai-global-inputs input').count(),0);
  await page.evaluate(data=>{authenticated=true;csrf='old-session';document.getElementById('workspace').hidden=false;document.getElementById('auth').hidden=true;aiState=data;renderAIGlobal(data);renderAIGroups(data.groups);aiBusyState(false);},fixture());
  await page.evaluate(()=>{void loadAI();});await page.waitForFunction(()=>pending.length===1);
  await page.evaluate(()=>{authenticated=false;csrf='';showAuth();authenticated=true;csrf='new-session';document.getElementById('workspace').hidden=false;document.getElementById('auth').hidden=true;void loadAI();});
  await page.waitForFunction(()=>pending.length===2);
  await page.evaluate(()=>pending[0]({ok:false,status:401,json:async()=>({error:'旧会话过期'})}));
  await page.waitForTimeout(30);
  assert.equal(await page.evaluate(()=>authenticated),true);assert.equal(await page.evaluate(()=>aiBusy),true);
  assert.equal(await page.locator('#reload-ai').isDisabled(),true);
  await page.evaluate(data=>pending[1]({ok:true,status:200,json:async()=>data}),fixture());await waitIdle(page);
  assert.equal(await page.evaluate(()=>aiState.global.fields.AgentChar),fixture().global.fields.AgentChar);
  assert.deepEqual(await dirty(page),[]);
  await page.locator('#ai-field-AgentChar').fill('过期草稿');
  await page.evaluate(()=>replies.push({error:'会话过期',status:401}));await page.locator('#save-ai').click();
  await page.waitForFunction(()=>!authenticated);assert.deepEqual(await dirty(page),[]);assert.equal(await page.evaluate(()=>aiBusy),false);
});

test('beforeunload prompts only while dirty or pending, and busy logout sends no request',async t=>{
  const page=await harness(t);
  const unload=()=>page.evaluate(()=>{const e=new Event('beforeunload',{cancelable:true});window.dispatchEvent(e);return e.defaultPrevented;});
  assert.equal(await unload(),false);await page.locator('#ai-field-AgentChar').fill('草稿');assert.equal(await unload(),true);
  await page.evaluate(()=>{void document.getElementById('ai-global-form').requestSubmit();});await page.waitForFunction(()=>pending.length===1);
  const dialog=page.waitForEvent('dialog');const click=page.locator('#logout').click();const d=await dialog;assert.equal(d.type(),'alert');await d.accept();await click;
  assert.equal(await page.evaluate(()=>requests.length),1);assert.equal(await unload(),true);
  await page.evaluate(data=>{const response=structuredClone(data);response.global.fields.AgentChar='草稿';pending[0]({ok:true,status:200,json:async()=>response});},fixture());await waitIdle(page);
  assert.equal(await unload(),false);
});

test('private scope reads an explicit target and saves the loaded conversation after editing the selector',async t=>{
  const page=await harness(t);const data=fixture();data.groups=data.groups.map(g=>({...g,groupId:'',privateId:'123456789'}));
  await page.selectOption('#ai-scope-kind','private');await page.locator('#ai-group').fill('123456789');
  await page.evaluate(data=>replies.push(data),data);await page.locator('#apply-ai-group').click();await waitIdle(page);
  assert.equal((await page.evaluate(()=>requests))[0].url,'/api/ai?private=123456789');
  assert.match(await page.locator('#ai-group-forms form[data-service="llm"] h2').textContent(),/GPT 温度 · 私聊 QQ 123456789/);
  await page.locator('#ai-field-temperature-llm').fill('81');
  await page.selectOption('#ai-scope-kind','group');await page.locator('#ai-group').fill('987654321');
  const saved=structuredClone(data);saved.groups[1].temperature=81;
  await page.evaluate(data=>replies.push(data),saved);await page.locator('#ai-group-forms form[data-service="llm"] button[type="submit"]').click();await waitIdle(page);
  const body=(await page.evaluate(()=>requests))[1].body;
  assert.equal(body.privateId,'123456789');assert.equal(body.groupId,undefined);assert.equal(body.temperature,81);
  await page.evaluate(data=>replies.push(data),saved);await page.locator('#reload-ai').click();await waitIdle(page);
  assert.equal((await page.evaluate(()=>requests))[2].url,'/api/ai?private=123456789');
  assert.equal(await page.locator('#ai-scope-kind').inputValue(),'private');
});

test('cancelled private switch or failed read preserves the loaded group and draft target',async t=>{
  const page=await harness(t);await page.locator('#ai-field-temperature-aichatcfg').fill('79');
  await page.selectOption('#ai-scope-kind','private');await page.locator('#ai-group').fill('123456789');
  await dialogAction(page,false,()=>page.locator('#apply-ai-group').click());assert.equal(await page.evaluate(()=>requests.length),0);
  await page.evaluate(()=>replies.push({error:'模拟私聊读取失败',status:500}));
  await dialogAction(page,true,()=>page.locator('#apply-ai-group').click());await waitIdle(page);
  assert.equal(await page.evaluate(()=>aiPrivate),'');assert.equal(await page.evaluate(()=>aiGroup),'123456789');
  assert.equal(await page.locator('#ai-field-temperature-aichatcfg').inputValue(),'79');
  assert.match(await page.locator('#ai-group-forms form[data-service="aichatcfg"] h2').textContent(),/群 123456789/);
  const saved=fixture();saved.groups[0].temperature=79;
  await page.evaluate(data=>replies.push(data),saved);await page.locator('#ai-group-forms form[data-service="aichatcfg"] button[type="submit"]').click();await waitIdle(page);
  const body=(await page.evaluate(()=>requests))[1].body;assert.equal(body.groupId,'123456789');assert.equal(body.privateId,undefined);
});
