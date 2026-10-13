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

const summary=(p,name)=>p.locator(`[data-ai-draft-form="${name}"]`);
const labels=(p,name)=>summary(p,name).locator('li').allTextContents();

test('all registered global fields and conversation controls have fixed labels; reverting or discarding only removes that form',async t=>{
 const p=await harness(t);assert.equal(await p.locator('#ai-draft-fields').isHidden(),true);
 await p.evaluate(()=>{for(const [name,,kind] of aiFields){const input=document.getElementById('ai-field-'+name);if(kind==='checkbox')input.checked=!input.checked;else input.value=kind==='type'?'1':kind==='number'?(name==='TopP'?'0.5':'8192'):'fixture-private-'+name+' <img src=x>';input.dispatchEvent(new Event('input',{bubbles:true}));}});
 assert.equal((await labels(p,'global')).length,17);assert.deepEqual(await labels(p,'global'),await p.evaluate(()=>aiFields.map(([,label])=>label)));assert.match(await summary(p,'global').textContent(),/未保存字段：17/);assert.doesNotMatch(await p.locator('#ai-draft-fields').innerHTML(),/fixture-private|<img|fixture.invalid|revision/);
 await p.locator('#ai-field-temperature-aichatcfg').fill('85');await p.locator('#ai-field-rate-aichatcfg').fill('20');await p.locator('#ai-field-agent-aichatcfg').uncheck();await p.locator('#ai-field-voice-aichatcfg').check();await p.locator('#ai-field-at-aichatcfg').uncheck();await p.locator('#ai-field-temperature-llm').fill('95');
 assert.deepEqual(await labels(p,'aichatcfg'),['温度百分比','非 @ 消息触发概率','Agent 模式','QQ AI 语音输出','响应 @ 消息']);assert.deepEqual(await labels(p,'llm'),['温度百分比']);assert.equal(await p.locator('[data-ai-draft-form]').count(),3);
 await p.locator('#ai-field-temperature-llm').fill('70');assert.equal(await summary(p,'llm').count(),0);await dialogAction(p,true,()=>p.locator('#discard-ai-global').click());assert.equal(await summary(p,'global').count(),0);assert.equal((await labels(p,'aichatcfg')).length,5);await dialogAction(p,true,()=>p.locator('#discard-ai-aichatcfg').click());assert.equal(await p.locator('#ai-draft-fields').isHidden(),true);assert.equal(await p.locator('#ai-draft-field-list').textContent(),'');assert.equal(await p.evaluate(()=>requests.length),0);
});

test('key actions, replacement values and clear confirmations merge per key; unknown input names and values never enter summary',async t=>{
 const p=await harness(t);await p.selectOption('#ai-action-Key','replace');await p.locator('#ai-key-Key').fill('fixture-private-replacement');assert.deepEqual(await labels(p,'global'),['聊天密钥']);await p.selectOption('#ai-action-AgentKey','clear');await p.locator('#ai-clear-AgentKey').check();assert.deepEqual(await labels(p,'global'),['聊天密钥','Agent 密钥']);
 await p.evaluate(()=>{const unknown=document.createElement('input');unknown.id='fixture-private-name';unknown.value='fixture-private-value';document.getElementById('ai-global-form').append(unknown);unknown.dispatchEvent(new Event('input',{bubbles:true}));});assert.deepEqual(await labels(p,'global'),['聊天密钥','Agent 密钥','其他表单字段']);assert.match(await summary(p,'global').textContent(),/未保存字段：3/);assert.doesNotMatch(await p.locator('#ai-draft-fields').innerHTML(),/fixture-private|replace|clear|keep|fixture.invalid/);
 await p.selectOption('#ai-action-Key','keep');assert.deepEqual(await labels(p,'global'),['Agent 密钥','其他表单字段']);await p.selectOption('#ai-action-AgentKey','keep');await p.evaluate(()=>{document.getElementById('fixture-private-name').remove();renderAIDrafts()});assert.equal(await p.locator('#ai-draft-fields').isHidden(),true);assert.equal(await p.evaluate(()=>requests.length),0);
});

test('private and group summaries use each form loaded target, ignoring selector draft and hiding unverifiable scope metadata',async t=>{
 const p=await harness(t);await p.locator('#ai-field-temperature-aichatcfg').fill('85');await p.locator('#ai-group').fill('345678901');await p.selectOption('#ai-scope-kind','private');assert.match(await summary(p,'aichatcfg').locator('h3').textContent(),/群 123456789/);assert.doesNotMatch(await summary(p,'aichatcfg').textContent(),/345678901/);
 const data=fixture();data.groups=data.groups.map(g=>({...g,groupId:'0',privateId:'234567890'}));await p.evaluate(value=>{aiState=value;aiGroup='0';aiPrivate='234567890';renderAIGroups(value.groups)},data);await p.locator('#ai-field-temperature-llm').fill('90');await p.locator('#ai-group').fill('345678901');assert.match(await summary(p,'llm').locator('h3').textContent(),/私聊 QQ 234567890/);assert.doesNotMatch(await summary(p,'llm').textContent(),/345678901/);
 await p.evaluate(value=>{value.groups=value.groups.map(g=>({...g,groupId:'234567890',privateId:''}));aiState=value;renderAIGroups(value.groups)},data);await p.locator('#ai-field-temperature-llm').fill('91');assert.match(await summary(p,'llm').locator('h3').textContent(),/群 234567890/);
 await p.evaluate(()=>{aiBaselines.get('llm').scope={privateId:'fixture-private-target',groupId:'0'};renderAIDrafts()});assert.match(await summary(p,'llm').locator('h3').textContent(),/范围暂不可核对/);assert.doesNotMatch(await p.locator('#ai-draft-fields').innerHTML(),/fixture-private-target/);assert.equal(await p.evaluate(()=>requests.length),0);
});

test('saving one form refreshes only its baseline; busy, conflict and failed reload retain summary and successful reload replaces it',async t=>{
 const p=await harness(t);await p.locator('#ai-field-AgentChar').fill('saved persona');await p.locator('#ai-field-temperature-aichatcfg').fill('80');await p.locator('#ai-field-temperature-llm').fill('90');await p.locator('#save-ai').click();await p.waitForFunction(()=>pending.length===1);assert.match(await p.locator('#ai-draft-field-state').textContent(),/请求处理中/);assert.equal(await p.locator('[data-ai-draft-form]').count(),3);
 const saved=fixture();saved.global.fields.AgentChar='saved persona';await p.evaluate(value=>pending[0]({ok:true,status:200,json:async()=>value}),saved);await waitIdle(p);assert.equal(await summary(p,'global').count(),0);assert.equal(await p.locator('[data-ai-draft-form]').count(),2);await p.locator('#ai-field-AgentChar').fill('different persona');assert.equal((await labels(p,'global')).length,1);await p.locator('#ai-field-AgentChar').fill('saved persona');assert.equal(await summary(p,'global').count(),0);
 await p.evaluate(()=>replies.push({error:'fixture conflict',status:409}));await p.locator('form[data-service="aichatcfg"] button[type="submit"]').click();await waitIdle(p);assert.equal(await p.locator('[data-ai-draft-form]').count(),2);const before=await p.locator('#ai-draft-field-list').textContent();await p.evaluate(()=>replies.push({error:'fixture read failed',status:503}));await dialogAction(p,true,()=>p.locator('#reload-ai').click());await waitIdle(p);assert.equal(await p.locator('#ai-draft-field-list').textContent(),before);
 await p.evaluate(value=>replies.push(value),saved);await dialogAction(p,true,()=>p.locator('#reload-ai').click());await waitIdle(p);assert.equal(await p.locator('#ai-draft-fields').isHidden(),true);assert.equal(await p.evaluate(()=>requests.length),4);
});

test('navigation preserves summary; reset/logout/current expiry clear it and stale success or 401 cannot refill fresh session fields',async t=>{
 for(const status of [200,401]){const p=await harness(t);await p.locator('#ai-field-AgentChar').fill('old draft');const before=await p.locator('#ai-draft-field-list').textContent();await p.locator('[data-view=overview]').click();await p.locator('[data-view=ai]').click();assert.equal(await p.locator('#ai-draft-field-list').textContent(),before);await dialogAction(p,true,()=>p.locator('#reload-ai').click());await p.waitForFunction(()=>pending.length===1);await p.evaluate(()=>{authenticated=false;csrf='';showAuth()});assert.equal(await p.locator('#ai-draft-fields').isHidden(),true);assert.equal(await p.locator('#ai-draft-field-list').textContent(),'');
 await p.evaluate(value=>{authenticated=true;csrf='fresh-session';aiState=value;document.getElementById('auth').hidden=true;document.getElementById('workspace').hidden=false;renderAIGlobal(value);renderAIGroups(value.groups);aiBusyState(false)},fixture());await p.locator('#ai-field-ModelName').fill('fixture-private-fresh');const fresh=await p.locator('#ai-draft-field-list').textContent();await p.evaluate(({status,value})=>pending[0]({ok:status===200,status,json:async()=>status===200?value:{error:'old expiry'}}),{status,value:fixture()});await p.waitForTimeout(30);assert.equal(await p.evaluate(()=>authenticated),true);assert.equal(await p.locator('#ai-draft-field-list').textContent(),fresh);assert.deepEqual(await labels(p,'global'),['聊天模型名']);assert.doesNotMatch(await p.locator('#ai-draft-fields').innerHTML(),/fixture-private-fresh/);
 await p.evaluate(()=>replies.push({error:'current expiry',status:401}));await p.locator('#save-ai').click();await p.waitForFunction(()=>!authenticated);assert.equal(await p.locator('#ai-draft-field-list').textContent(),'');assert.equal(await p.locator('#ai-draft-fields').isHidden(),true);}
 const p=await harness(t);await p.locator('#ai-field-AgentChar').fill('restore draft');await p.evaluate(()=>resetAI());assert.equal(await p.locator('#ai-draft-fields').isHidden(),true);assert.equal(await p.locator('#ai-draft-field-list').textContent(),'');assert.equal(await p.evaluate(()=>requests.length),0);
});

