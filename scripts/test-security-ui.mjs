// Offline Chrome with synthetic passwords and mocked API, never the user's browser/password.
import assert from 'node:assert/strict';
import fs from 'node:fs/promises';
import path from 'node:path';
import {createRequire} from 'node:module';
import test from 'node:test';
const require=createRequire(import.meta.url);
const modules=process.env.LOCALADMIN_NODE_MODULES || path.join(process.env.USERPROFILE,'.cache/codex-runtimes/codex-primary-runtime/dependencies/node/node_modules');const {chromium}=require(path.join(modules,'playwright'));
const html=(await fs.readFile('internal/localadmin/assets/index.html','utf8')).replace(/<script\b[^>]*>[\s\S]*?<\/script>/g,'').replace(/<link\b[^>]*>/g,'');
const app=await fs.readFile('internal/localadmin/assets/app.js','utf8'),security=await fs.readFile('internal/localadmin/assets/security.js','utf8');
async function fixture(t){const browser=await chromium.launch({channel:'chrome',headless:true});t.after(()=>browser.close());const context=await browser.newContext();await context.route('**/*',r=>r.abort());const page=await context.newPage();await page.setContent(html);await page.addScriptTag({content:app});await page.addScriptTag({content:security});await page.evaluate(()=>{
 authenticated=true;csrf='security-fixture';window.showOverview=()=>{};window.showBackups=()=>{};window.showLogs=()=>{};
 window.requests=[];window.pending=[];window.replies=[];window.confirmations=[];window.confirmAnswer=true;window.storageWrites=0;
 Storage.prototype.setItem=()=>storageWrites++;
 window.confirm=text=>{confirmations.push(text);return confirmAnswer};window.fetch=(url,options)=>{requests.push({url,body:options.body?JSON.parse(options.body):null,headers:options.headers,signal:options.signal});const reply=replies.shift();if(reply){if(reply.network)return Promise.reject(new TypeError('synthetic connection lost'));return Promise.resolve({ok:!reply.error,status:reply.status||200,json:async()=>reply});}return new Promise(resolve=>pending.push(resolve));};
 document.getElementById('auth').hidden=true;document.getElementById('workspace').hidden=false;document.querySelector('.nav-button[data-view=security]').click();
 });return page;}
const old='fixture-old-password-only',next='fixture-new-password-only';
async function fill(page,current=old,newValue=next,confirm=newValue){await page.locator('#current-admin-password').fill(current);await page.locator('#new-admin-password').fill(newValue);await page.locator('#confirm-admin-password').fill(confirm);}
async function submit(page){await page.locator('#change-admin-password').click();}
async function blank(page){for(const id of ['current-admin-password','new-admin-password','confirm-admin-password'])assert.equal(await page.locator('#'+id).inputValue(),'');}

test('unicode/byte/confirmation validation, cancel and successful change use actual app logout with no password storage',async t=>{
 const page=await fixture(t);assert.match(await page.locator('#breadcrumb').textContent(),/后台安全/);
 for(const pair of [[next,'mismatched-password'],[old,old],['🙂'.repeat(6),'🙂'.repeat(6)],['汉'.repeat(171),'汉'.repeat(171)]]) {await fill(page,old,...pair);await submit(page);assert.equal(await page.locator('#security-error').isVisible(),true);assert.equal(await page.evaluate(()=>requests.length),0);}
 await fill(page);await page.evaluate(()=>confirmAnswer=false);await submit(page);assert.equal(await page.locator('#new-admin-password').inputValue(),next);assert.equal(await page.evaluate(()=>requests.length),0);
 await page.evaluate(()=>{confirmAnswer=true;replies.push({passwordChanged:true,authenticated:false});pluginDrafts.set('fixture','disabled');});await submit(page);await page.waitForFunction(()=>!authenticated);
 assert.equal(await page.locator('#auth').isVisible(),true);assert.match(await page.locator('#auth-notice').textContent(),/密码已修改/);await blank(page);assert.equal(await page.evaluate(()=>pluginDrafts.size),0);
 const requests=await page.evaluate(()=>requests);assert.equal(requests.length,1);assert.equal(requests[0].url,'/api/password');assert.equal(requests[0].headers['X-CSRF-Token'],'security-fixture');assert.deepEqual(requests[0].body,{currentPassword:old,newPassword:next,confirmPassword:next});
 assert.equal(await page.evaluate(()=>storageWrites),0);assert.equal((await page.locator('body').textContent()).includes(next),false);assert.equal((await page.evaluate(()=>confirmations)).some(text=>text.includes(old)||text.includes(next)),false);
});

test('wrong current password, throttle, write failure and network failure clear input without falsely logging out',async t=>{
 const page=await fixture(t);
 for(const reply of [{error:'当前后台密码不正确',status:403},{error:'一分钟后重试',status:429},{error:'密码保存失败，旧密码保留',status:500},{network:true}]){
 await fill(page);await page.evaluate(reply=>replies.push(reply),reply);await submit(page);await page.waitForFunction(()=>!securityBusy);await blank(page);assert.equal(await page.evaluate(()=>authenticated),true);assert.equal(await page.locator('#security-error').isVisible(),true);assert.equal(await page.locator('#auth-notice').textContent(),'');
 }
});

test('leaving clears password inputs; pending request freezes navigation/logout/clear and warns on departure',async t=>{
 const page=await fixture(t);await fill(page);await page.locator('.nav-button[data-view=overview]').click();await blank(page);await page.locator('.nav-button[data-view=security]').click();await fill(page);
 await page.evaluate(()=>connectionBusy=true);await submit(page);assert.equal(await page.evaluate(()=>requests.length),0);assert.match(await page.locator('#security-error').textContent(),/其他设置请求/);assert.equal(await page.locator('#new-admin-password').inputValue(),next);
 await page.evaluate(()=>connectionBusy=false);await submit(page);await page.waitForFunction(()=>securityBusy);assert.equal(await page.locator('#logout').isDisabled(),true);assert.equal(await page.locator('#clear-admin-passwords').isDisabled(),true);
 await page.evaluate(()=>{document.querySelector('.nav-button[data-view=overview]').dispatchEvent(new Event('click'));document.getElementById('logout').dispatchEvent(new Event('click'));});assert.equal(await page.locator('#security-view').isVisible(),true);assert.equal(await page.evaluate(()=>requests.length),1);
 assert.equal(await page.evaluate(()=>{const event=new Event('beforeunload',{cancelable:true});window.dispatchEvent(event);return event.defaultPrevented;}),true);
 await page.evaluate(()=>pending[0]({ok:false,status:403,json:async()=>({error:'wrong current'})}));await page.waitForFunction(()=>!securityBusy);await blank(page);assert.equal(await page.locator('#logout').isEnabled(),true);
 await fill(page);await page.locator('#clear-admin-passwords').click();await blank(page);
});

test('late success and old 401 cannot clear or unlock new login; current expiry clears secrets',async t=>{
 for(const status of [200,401]){
 const page=await fixture(t);await fill(page);await submit(page);await page.waitForFunction(()=>pending.length===1);
 await page.evaluate(()=>{authenticated=false;showAuth();authenticated=true;csrf='new-security-fixture';document.getElementById('auth').hidden=true;document.getElementById('workspace').hidden=false;showSecurity(true);});await fill(page,'other-old-password','other-new-password');await submit(page);await page.waitForFunction(()=>pending.length===2);
 await page.evaluate(status=>pending[0]({ok:status===200,status,json:async()=>status===200?{passwordChanged:true,authenticated:false}:{error:'old expired'}}),status);await page.evaluate(()=>Promise.resolve());assert.equal(await page.evaluate(()=>authenticated),true);assert.equal(await page.evaluate(()=>securityBusy),true);assert.equal(await page.evaluate(()=>csrf),'new-security-fixture');assert.equal(await page.locator('#new-admin-password').inputValue(),'other-new-password');
 await page.evaluate(()=>pending[1]({ok:false,status:401,json:async()=>({error:'current expired'})}));await page.waitForFunction(()=>!authenticated);await blank(page);assert.equal(await page.evaluate(()=>securityBusy),false);
 }
});

test('stale confirmation, duplicate submit and unexpected success payload never report a successful change',async t=>{
 const page=await fixture(t);await fill(page);await page.evaluate(()=>window.confirm=()=>{resetSecurity();showSecurity(true);return true;});await submit(page);await blank(page);assert.equal(await page.evaluate(()=>requests.length),0);
 await fill(page);await page.evaluate(()=>window.confirm=()=>true);await submit(page);await page.waitForFunction(()=>pending.length===1);await page.evaluate(()=>document.getElementById('password-change-form').dispatchEvent(new Event('submit',{cancelable:true})));assert.equal(await page.evaluate(()=>requests.length),1);
 await page.evaluate(()=>pending[0]({ok:true,status:200,json:async()=>({authenticated:true})}));await page.waitForFunction(()=>!securityBusy);await blank(page);assert.equal(await page.evaluate(()=>authenticated),true);assert.match(await page.locator('#security-error').textContent(),/提交结果无法确认/);assert.equal(await page.locator('#auth-notice').textContent(),'');
});
