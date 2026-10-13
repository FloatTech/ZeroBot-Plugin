// Real Chrome navigation with synthetic logs only; no live browser sessions or network.
import assert from 'node:assert/strict';
import fs from 'node:fs/promises';
import path from 'node:path';
import {createRequire} from 'node:module';
import test from 'node:test';
const require=createRequire(import.meta.url);
const modules=process.env.LOCALADMIN_NODE_MODULES || path.join(process.env.USERPROFILE,'.cache/codex-runtimes/codex-primary-runtime/dependencies/node/node_modules');
const {chromium}=require(path.join(modules,'playwright'));
const html=(await fs.readFile('internal/localadmin/assets/index.html','utf8')).replace(/<script\b[^>]*>[\s\S]*?<\/script>/g,'').replace(/<link\b[^>]*>/g,'');
const app=await fs.readFile('internal/localadmin/assets/app.js','utf8');
const logs=await fs.readFile('internal/localadmin/assets/logs.js','utf8');

test('overview entry navigates through actual app to applied issues, preserving pause/autoscroll and exporting same filter',async t=>{
  const browser=await chromium.launch({channel:'chrome',headless:true});t.after(()=>browser.close());
  const context=await browser.newContext();await context.route('**/*',r=>r.abort());
  const page=await context.newPage();await page.setContent(html);await page.addScriptTag({content:app});await page.addScriptTag({content:logs});
  await page.evaluate(()=>{
    authenticated=true;csrf='fixture-session';window.showOverview=()=>{};window.showBackups=()=>{};
    document.getElementById('auth').hidden=true;document.getElementById('workspace').hidden=false;
    window.requests=[];window.fetch=async(url)=>{requests.push(url);return {ok:true,status:200,json:async()=>({entries:[{id:'1',time:'2026-10-09T00:00:00Z',level:'warning',message:'synthetic warning'}],cursor:'1',instance:'0123456789abcdef',retained:1,capacity:1000})};};
    window.downloads=[];URL.createObjectURL=()=> 'blob:fixture';URL.revokeObjectURL=()=>{};
    HTMLAnchorElement.prototype.click=function(){downloads.push(this.download);};
    logLevel='info';logKeyword='previous';document.getElementById('log-level').value='error';document.getElementById('log-keyword').value='unapplied';
    logPaused=true;document.getElementById('pause-logs').textContent='恢复更新';document.getElementById('log-autoscroll').checked=false;
  });
  await page.locator('#view-issue-logs').click();
  assert.equal(await page.locator('.nav-button.active').getAttribute('data-view'),'logs');assert.equal(await page.locator('#logs-view').isVisible(),true);
  assert.equal(await page.locator('#log-level').inputValue(),'issues');assert.equal(await page.locator('#log-keyword').inputValue(),'');
  assert.match(await page.locator('#log-filter-status').textContent(),/警告与错误/);assert.doesNotMatch(await page.locator('#log-filter-status').textContent(),/尚未应用/);
  assert.equal(await page.evaluate(()=>logPaused),true);assert.equal(await page.locator('#log-autoscroll').isChecked(),false);assert.equal(await page.evaluate(()=>requests.length),0);
  await page.locator('#pause-logs').click();await page.waitForFunction(()=>document.querySelectorAll('#log-records article').length===1);
  let query=new URL((await page.evaluate(()=>requests))[0],'http://fixture.invalid').searchParams;
  assert.equal(query.get('level'),'issues');assert.equal(query.get('q'),'');assert.equal(query.get('after'),'0');
  await page.locator('#log-keyword').fill('unapplied draft');assert.match(await page.locator('#log-filter-status').textContent(),/尚未应用/);
  await page.evaluate(()=>{window.fetch=async(url)=>{requests.push(url);return {ok:true,status:200,headers:new Headers({'Content-Type':'text/plain; charset=utf-8','Content-Length':'7','Content-Disposition':'attachment; filename="zerobot-logs-20261009T000000Z.txt"'}),blob:async()=>new Blob(['fixture'])};};});
  await page.locator('#export-logs').click();await page.waitForFunction(()=>downloads.length===1);
  query=new URL((await page.evaluate(()=>requests)).at(-1),'http://fixture.invalid').searchParams;
  assert.equal(query.get('level'),'issues');assert.equal(query.get('q'),'');
  await page.locator('#pause-logs').click();await page.locator('#clear-log-filter').click();
  assert.equal(await page.locator('#log-level').inputValue(),'');assert.equal(await page.locator('#log-keyword').inputValue(),'');assert.equal(await page.evaluate(()=>logPaused),true);
});
