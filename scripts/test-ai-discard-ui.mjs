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

test('confirmed discard restores only the chosen form, while cancellation keeps all three drafts and sends nothing',async t=>{
  const page=await harness(t);
  for(const name of ['global','aichatcfg','llm'])assert.equal(await page.locator('#discard-ai-'+name).isDisabled(),true);
  await page.locator('#ai-field-AgentChar').fill('global draft');await page.locator('#ai-field-temperature-aichatcfg').fill('80');await page.locator('#ai-field-temperature-llm').fill('90');
  await page.locator('#ai-field-rate-aichatcfg').fill('25');await page.locator('#ai-field-agent-aichatcfg').uncheck();await page.locator('#ai-field-voice-aichatcfg').check();
  await dialogAction(page,false,()=>page.locator('#discard-ai-aichatcfg').click());assert.deepEqual(await dirty(page),['global','aichatcfg','llm']);assert.equal(await page.locator('#ai-field-temperature-aichatcfg').inputValue(),'80');
  await dialogAction(page,true,()=>page.locator('#discard-ai-aichatcfg').click());assert.deepEqual(await dirty(page),['global','llm']);assert.equal(await page.locator('#ai-field-temperature-aichatcfg').inputValue(),'70');assert.equal(await page.locator('#ai-field-rate-aichatcfg').inputValue(),'0');assert.equal(await page.locator('#ai-field-agent-aichatcfg').isChecked(),true);assert.equal(await page.locator('#ai-field-voice-aichatcfg').isChecked(),false);
  assert.equal(await page.locator('#ai-field-AgentChar').inputValue(),'global draft');assert.equal(await page.locator('#ai-field-temperature-llm').inputValue(),'90');
  await dialogAction(page,true,()=>page.locator('#discard-ai-llm').click());assert.deepEqual(await dirty(page),['global']);
  await dialogAction(page,true,()=>page.locator('#discard-ai-global').click());assert.deepEqual(await dirty(page),[]);assert.equal(await page.evaluate(()=>requests.length),0);
});

test('global discard resets field types and wipes unsaved replacement keys, keeping saved-key flags and other drafts',async t=>{
  const page=await harness(t);await page.locator('#ai-field-temperature-llm').fill('90');
  await page.selectOption('#ai-field-Type','1');await page.locator('#ai-field-MaxN').fill('8192');await page.locator('#ai-field-TopP').fill('0.5');await page.locator('#ai-field-NoSystemP').check();await page.locator('#ai-field-SystemP').fill('<script>synthetic prompt</script>');
  for(const key of ['Key','ImageKey']){await page.selectOption('#ai-action-'+key,'replace');await page.locator('#ai-key-'+key).fill('synthetic-private-'+key);}
  await page.selectOption('#ai-action-AgentKey','clear');await page.locator('#ai-clear-AgentKey').check();
  let message='';page.on('dialog',dialog=>{message=dialog.message();});
  await dialogAction(page,true,()=>page.locator('#discard-ai-global').click());assert.doesNotMatch(message,/synthetic-private|synthetic prompt/);
  for(const key of ['Key','ImageKey','AgentKey']){
    assert.equal(await page.locator('#ai-action-'+key).inputValue(),'keep');assert.equal(await page.locator('#ai-key-'+key).inputValue(),'');assert.equal(await page.locator('#ai-key-'+key).isDisabled(),true);
    assert.equal(await page.locator('#ai-key-'+key).evaluate(el=>el.required),false);assert.equal(await page.locator('#ai-clear-'+key).isChecked(),false);assert.equal(await page.locator('#ai-clear-'+key).locator('..').isHidden(),true);
  }
  assert.equal(await page.locator('#ai-field-Type').inputValue(),'0');assert.equal(await page.locator('#ai-field-MaxN').inputValue(),'4096');assert.equal(await page.locator('#ai-field-TopP').inputValue(),'0.9');assert.equal(await page.locator('#ai-field-NoSystemP').isChecked(),false);assert.equal(await page.locator('#ai-field-SystemP').inputValue(),'fixture prompt');
  assert.deepEqual(await page.evaluate(()=>aiState.global.keys),fixture().global.keys);assert.deepEqual(await dirty(page),['llm']);assert.equal(await page.locator('#ai-field-temperature-llm').inputValue(),'90');assert.equal(await page.evaluate(()=>requests.length),0);
  assert.doesNotMatch(ai,/localStorage|sessionStorage|indexedDB/);
});

test('discard uses the last successful baseline, preserves other forms, and retains the version after conflicts',async t=>{
  const page=await harness(t);await page.locator('#ai-field-AgentChar').fill('saved persona');
  const saved=fixture();saved.global.fields.AgentChar='saved persona';saved.global.revision='c'.repeat(64);
  await page.evaluate(value=>replies.push(value),saved);await page.locator('#save-ai').click();await waitIdle(page);
  await page.locator('#ai-field-AgentChar').fill('unsaved persona');await page.locator('#ai-field-temperature-aichatcfg').fill('80');await page.locator('#ai-field-temperature-llm').fill('90');
  await dialogAction(page,true,()=>page.locator('#discard-ai-global').click());assert.equal(await page.locator('#ai-field-AgentChar').inputValue(),'saved persona');assert.equal(await page.evaluate(()=>aiState.global.revision),'c'.repeat(64));assert.deepEqual(await dirty(page),['aichatcfg','llm']);
  await page.locator('#ai-field-AgentChar').fill('keep global draft');await page.evaluate(groups=>replies.push({groups}),[group('aichatcfg',80),group('llm')]);
  await page.locator('form[data-service="aichatcfg"] button[type="submit"]').click();await waitIdle(page);
  await page.locator('#ai-field-temperature-aichatcfg').fill('82');await dialogAction(page,true,()=>page.locator('#discard-ai-aichatcfg').click());assert.equal(await page.locator('#ai-field-temperature-aichatcfg').inputValue(),'80');assert.equal(await page.locator('#ai-field-temperature-llm').inputValue(),'90');assert.equal(await page.locator('#ai-field-AgentChar').inputValue(),'keep global draft');
  await page.evaluate(()=>replies.push({error:'版本冲突',status:409}));await page.locator('#save-ai').click();await waitIdle(page);assert.deepEqual(await dirty(page),['global','llm']);
  await dialogAction(page,true,()=>page.locator('#discard-ai-global').click());assert.equal(await page.locator('#ai-field-AgentChar').inputValue(),'saved persona');assert.equal(await page.evaluate(()=>aiState.global.revision),'c'.repeat(64));assert.equal(await page.evaluate(()=>requests.length),3);assert.equal(await page.evaluate(()=>requests[2].body.revision),'c'.repeat(64));
});

test('private form discard respects the loaded target and preserves the uncommitted selector before a later save',async t=>{
  const page=await harness(t),data=fixture();data.groups=data.groups.map(g=>({...g,groupId:'0',privateId:'234567890'}));
  await page.evaluate(value=>{aiState=value;aiGroup='0';aiPrivate='234567890';document.getElementById('ai-scope-kind').value='private';document.getElementById('ai-group').value='234567890';renderAIScopeLabel();renderAIGroups(value.groups);},data);
  await page.locator('#ai-group').fill('345678901');await page.locator('#ai-field-temperature-aichatcfg').fill('85');await page.locator('#ai-field-temperature-llm').fill('95');
  let message='';page.on('dialog',dialog=>{message=dialog.message();});await dialogAction(page,true,()=>page.locator('#discard-ai-aichatcfg').click());assert.match(message,/私聊 QQ 234567890/);assert.doesNotMatch(message,/345678901/);
  assert.equal(await page.locator('#ai-group').inputValue(),'345678901');assert.equal(await page.evaluate(()=>aiPrivate),'234567890');assert.equal(await page.locator('#ai-field-temperature-aichatcfg').inputValue(),'70');assert.equal(await page.locator('#ai-field-temperature-llm').inputValue(),'95');assert.equal(await page.evaluate(()=>requests.length),0);
  await page.locator('#ai-field-temperature-aichatcfg').fill('83');data.groups[0].temperature=83;await page.evaluate(value=>replies.push({groups:value.groups}),data);
  await page.locator('form[data-service="aichatcfg"] button[type="submit"]').click();await waitIdle(page);assert.equal(await page.evaluate(()=>requests[0].body.privateId),'234567890');assert.equal(await page.evaluate(()=>Object.hasOwn(requests[0].body,'groupId')),false);
  assert.deepEqual(await dirty(page),['llm']);await dialogAction(page,true,()=>page.locator('#discard-ai-llm').click());assert.equal(await page.locator('#ai-field-temperature-llm').inputValue(),'70');
});

test('busy operations, expired sessions and late confirmations cannot restore or unlock newer form inputs',async t=>{
  const page=await harness(t);await page.locator('#ai-field-AgentChar').fill('busy draft');await page.locator('#ai-field-temperature-aichatcfg').fill('80');await page.locator('#ai-field-temperature-llm').fill('90');
  await page.locator('#save-ai').click();await page.waitForFunction(()=>pending.length===1);
  for(const name of ['global','aichatcfg','llm'])assert.equal(await page.locator('#discard-ai-'+name).isDisabled(),true);
  await page.evaluate(()=>discardAIForm('global'));assert.equal(await page.locator('#ai-field-AgentChar').inputValue(),'busy draft');
  const saved=fixture();saved.global.fields.AgentChar='busy draft';await page.evaluate(value=>pending[0]({ok:true,status:200,json:async()=>value}),saved);await waitIdle(page);
  await dialogAction(page,true,()=>page.locator('#reload-ai').click());await page.waitForFunction(()=>pending.length===2);assert.equal(await page.locator('#discard-ai-aichatcfg').isDisabled(),true);
  await page.evaluate(value=>{authenticated=false;csrf='';showAuth();authenticated=true;csrf='new-session';aiState=value;aiGroup='123456789';document.getElementById('auth').hidden=true;document.getElementById('workspace').hidden=false;renderAIGlobal(value);renderAIGroups(value.groups);aiBusyState(false);},fixture());
  await page.locator('#ai-field-AgentChar').fill('new session draft');await page.evaluate(()=>pending[1]({ok:false,status:401,json:async()=>({error:'old expiry'})}));await page.waitForTimeout(30);
  assert.equal(await page.evaluate(()=>authenticated),true);assert.equal(await page.locator('#ai-field-AgentChar').inputValue(),'new session draft');assert.equal(await page.locator('#discard-ai-global').isDisabled(),false);
  await page.evaluate(value=>{
    const confirm=window.confirm;
    window.confirm=()=>{resetAI();aiState=value;aiGroup='123456789';renderAIGlobal(value);renderAIGroups(value.groups);aiBusyState(false);document.getElementById('ai-field-AgentChar').value='after stale confirmation';renderAIDrafts();return true;};
    try{discardAIForm('global');}finally{window.confirm=confirm;}
  },fixture());
  assert.equal(await page.locator('#ai-field-AgentChar').inputValue(),'after stale confirmation');assert.deepEqual(await dirty(page),['global']);
  await page.evaluate(()=>replies.push({error:'session expired',status:401}));await page.locator('#save-ai').click();await page.waitForFunction(()=>!authenticated);assert.equal(await page.locator('#ai-global-inputs').textContent(),'');assert.equal(await page.locator('#discard-ai-global').isDisabled(),true);assert.deepEqual(await dirty(page),[]);
  await page.evaluate(value=>{authenticated=true;csrf='last-session';aiState=value;aiGroup='123456789';document.getElementById('auth').hidden=true;document.getElementById('workspace').hidden=false;renderAIGlobal(value);renderAIGroups(value.groups);aiBusyState(false);replies.push({});},fixture());
  await page.locator('#ai-field-AgentChar').fill('logout draft');await dialogAction(page,true,()=>page.locator('#logout').click());await page.waitForFunction(()=>!authenticated);assert.equal(await page.locator('#discard-ai-global').isDisabled(),true);assert.deepEqual(await dirty(page),[]);
});
