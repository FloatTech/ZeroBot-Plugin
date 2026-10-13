// Offline Chrome with synthetic registered usage only; no live site, clipboard or QQ.
import assert from 'node:assert/strict';
import fs from 'node:fs/promises';
import path from 'node:path';
import {createRequire} from 'node:module';
import test,{after} from 'node:test';
const require=createRequire(import.meta.url);
const modules=process.env.LOCALADMIN_NODE_MODULES||path.join(process.env.USERPROFILE,'.cache/codex-runtimes/codex-primary-runtime/dependencies/node/node_modules');
const {chromium}=require(path.join(modules,'playwright'));
const browser=await chromium.launch({channel:'chrome',headless:true});after(()=>browser.close());
const html=(await fs.readFile('internal/localadmin/assets/index.html','utf8')).replace(/<script\b[^>]*>[\s\S]*?<\/script>/g,'').replace(/<link\b[^>]*>/g,'');
const app=await fs.readFile('internal/localadmin/assets/app.js','utf8');
const css=await fs.readFile('internal/localadmin/assets/style.css','utf8')+'\n'+await fs.readFile('internal/localadmin/assets/appearance.css','utf8');
const usage='/示例 中文🙂\n第二行 <img src=x onerror=bad()>\n保留换行\n';
const plugin={name:'CON:/测试\\?"插件🙂',brief:'中文简介',usage,Token:'fixture-secret',groupId:'987654321',state:{effectiveEnabled:true,source:'default'}};
async function fixture(t){
 const context=await browser.newContext({acceptDownloads:true});t.after(()=>context.close());await context.route('**/*',r=>r.abort());
 const p=await context.newPage();await p.setContent(html);await p.addStyleTag({content:css});await p.addScriptTag({content:app});
 await p.evaluate(()=>{authenticated=true;csrf='usage-session';window.requests=[];window.copied=[];window.copyPending=[];window.clipboardMode='success';window.revoked=[];const revoke=URL.revokeObjectURL.bind(URL);URL.revokeObjectURL=u=>{revoked.push(u);revoke(u)};Object.defineProperty(navigator,'clipboard',{configurable:true,value:{writeText:async text=>{copied.push(text);if(clipboardMode==='fail')throw Error('private-system-error');if(clipboardMode==='pending')return new Promise((resolve,reject)=>copyPending.push({resolve,reject}));}}});window.fetch=async(url,options)=>{requests.push({url,method:options.method||'GET'});throw Error('Unexpected API request')};window.showOverview=()=>{};window.showLogs=()=>{};window.showBackups=()=>{};document.getElementById('auth').hidden=true;document.getElementById('workspace').hidden=false;document.getElementById('overview-view').hidden=true;document.getElementById('plugins-view').hidden=false;});
 return p;
}
const open=(p,data=plugin)=>p.evaluate(data=>openPluginUsage(data),data);
const status=p=>p.locator('#usage-export-status').textContent();

test('download is UTF-8 registered snapshot with a safe filename; no config, scope, drafts or HTML execution',async t=>{
 const p=await fixture(t);await open(p);
 await p.evaluate(()=>{plugins=[{name:'new-plugin',usage:'other'}];pluginGroup='123456789';pluginDrafts.set('new-plugin','disabled');renderPlugins();});
 assert.equal(await p.locator('#usage-text').textContent(),usage);assert.equal(await p.locator('#usage-text img').count(),0);
 const [download]=await Promise.all([p.waitForEvent('download'),p.locator('#download-plugin-usage').click()]);
 const text=await fs.readFile(await download.path(),'utf8');assert.equal(text,`插件：${plugin.name}\n简介：${plugin.brief}\n\n注册用法：\n${usage}\n`);assert.doesNotMatch(text,/fixture-secret|987654321|123456789|effectiveEnabled/);
 assert.match(download.suggestedFilename(),/^zerobot-plugin-CON__测试__/);assert.doesNotMatch(download.suggestedFilename(),/[<>:"/\\|?*\u0000-\u001f]/);
 assert.match(await status(p),/已发起文本下载/);await p.waitForFunction(()=>revoked.length===1);assert.equal(await p.locator('a[download]').count(),0);assert.equal(await p.evaluate(()=>requests.length),0);assert.equal(await p.evaluate(()=>pluginDrafts.size),1);
 await open(p,{name:'🙂'.repeat(100)+'... ',brief:'long',usage:'/long'});const [long]=await Promise.all([p.waitForEvent('download'),p.locator('#download-plugin-usage').click()]);assert.equal(Array.from(long.suggestedFilename().slice('zerobot-plugin-'.length,-4)).length,64);assert.doesNotMatch(long.suggestedFilename(),/\uFFFD/);
});

test('copy uses exact registered text, blocks duplicates and offers manual selection on failure or unavailable clipboard',async t=>{
 const p=await fixture(t);await open(p);await p.locator('#copy-plugin-usage').click();await p.waitForFunction(()=>!pluginUsage.copying);assert.match(await status(p),/已复制注册用法/);assert.deepEqual(await p.evaluate(()=>copied),[usage]);
 await p.evaluate(()=>clipboardMode='pending');await p.locator('#copy-plugin-usage').click();await p.evaluate(()=>copyPluginUsage());assert.equal(await p.locator('#copy-plugin-usage').isDisabled(),true);assert.equal(await p.evaluate(()=>copied.length),2);await p.evaluate(()=>copyPending[0].resolve());await p.waitForFunction(()=>!pluginUsage.copying);
 for(const unavailable of [false,true]){await p.evaluate(unavailable=>{clipboardMode='fail';if(unavailable)Object.defineProperty(navigator,'clipboard',{value:undefined,configurable:true});},unavailable);await p.locator('#copy-plugin-usage').click();await p.waitForFunction(()=>!pluginUsage.copying);assert.match(await status(p),/Ctrl\+C/);assert.doesNotMatch(await status(p),/private-system-error/);assert.equal(await p.evaluate(()=>getSelection().getRangeAt(0).toString()),usage);assert.equal(await p.evaluate(()=>getSelection().toString()),usage.trimEnd());assert.equal(await p.locator('#copy-plugin-usage').isEnabled(),true);}
 await p.locator('#close-dialog').click();assert.equal(await p.evaluate(()=>getSelection().toString()),'');assert.equal(await p.evaluate(()=>requests.length),0);
});

test('missing or invalid usage cannot copy or download placeholders; download errors do not leak details',async t=>{
 const p=await fixture(t);for(const value of [undefined,'',' \n\t ',42]){await open(p,{name:'empty',usage:value});assert.match(await p.locator('#usage-text').textContent(),/暂未提供/);assert.equal(await p.locator('#copy-plugin-usage').isDisabled(),true);assert.equal(await p.locator('#download-plugin-usage').isDisabled(),true);await p.evaluate(()=>{copyPluginUsage();downloadPluginUsage()});}
 assert.equal(await p.evaluate(()=>copied.length),0);await open(p);await p.evaluate(()=>URL.createObjectURL=()=>{throw Error('private-file-error')});await p.locator('#download-plugin-usage').click();assert.match(await status(p),/下载未完成/);assert.doesNotMatch(await status(p),/private-file-error/);assert.equal(await p.locator('#download-plugin-usage').isEnabled(),true);assert.equal(await p.evaluate(()=>requests.length),0);
});

test('closing, Escape and expiry clear content; late clipboard completion cannot overwrite another dialog or session',async t=>{
 for(const reject of [false,true]){const p=await fixture(t);await open(p);await p.evaluate(()=>clipboardMode='pending');await p.locator('#copy-plugin-usage').click();await p.locator('#close-dialog').click();assert.equal(await p.locator('#usage-title').textContent(),'');assert.equal(await p.locator('#usage-text').textContent(),'');await open(p,{name:'second',usage:'/second'});await p.evaluate(reject=>reject?copyPending[0].reject(Error('late-private-error')):copyPending[0].resolve(),reject);await p.waitForTimeout(40);assert.equal(await p.locator('#usage-title').textContent(),'second');assert.equal(await status(p),'');assert.equal(await p.locator('#copy-plugin-usage').isEnabled(),true);await p.keyboard.press('Escape');await p.waitForFunction(()=>pluginUsage===null);assert.equal(await p.locator('#usage-text').textContent(),'');
 await open(p);await p.locator('#copy-plugin-usage').click();await p.evaluate(()=>{authenticated=false;csrf='';showAuth()});assert.equal(await p.locator('#usage-dialog').evaluate(el=>el.open),false);assert.equal(await p.locator('#usage-text').textContent(),'');await p.evaluate(()=>{authenticated=true;csrf='new-usage-session';document.getElementById('auth').hidden=true;document.getElementById('workspace').hidden=false});await open(p,{name:'new-session',usage:'/new'});await p.evaluate(()=>copyPending[1].resolve());await p.waitForTimeout(40);assert.equal(await status(p),'');assert.equal(await p.locator('#usage-title').textContent(),'new-session');assert.equal(await p.locator('#copy-plugin-usage').isEnabled(),true);await p.evaluate(()=>{authenticated=false;csrf='';showAuth();openPluginUsage({name:'blocked',usage:'/blocked'})});assert.equal(await p.locator('#usage-dialog').evaluate(el=>el.open),false);assert.equal(await p.evaluate(()=>requests.length),0);}
});

test('usage dialog and actions fit four widths and both themes with long names and text',async t=>{
 const p=await fixture(t);for(const width of [320,375,640,1280]){await p.setViewportSize({width,height:900});for(const mode of ['light','dark']){await p.evaluate(mode=>document.documentElement.dataset.theme=mode,mode);await open(p,{name:'长插件🙂'.repeat(80),brief:'说明',usage:'长说明中文🙂'.repeat(800)+'\n/end'});const size=await p.evaluate(()=>{const dialog=document.getElementById('usage-dialog'),r=dialog.getBoundingClientRect();return {left:r.left,right:r.right,top:r.top,bottom:r.bottom,width:innerWidth,height:innerHeight,scrollWidth:dialog.scrollWidth,clientWidth:dialog.clientWidth}});assert.ok(size.left>=0&&size.right<=size.width+1&&size.top>=0&&size.bottom<=size.height+1,JSON.stringify(size));assert.ok(size.scrollWidth<=size.clientWidth+1,JSON.stringify(size));assert.equal(await p.locator('#copy-plugin-usage').isEnabled(),true);for(const id of ['copy-plugin-usage','download-plugin-usage','close-dialog']){const r=await p.locator('#'+id).boundingBox();assert.ok(r.y>=0&&r.y+r.height<=900,JSON.stringify({id,...r}));}if((width===1280&&mode==='light')||(width===375&&mode==='dark')){await fs.mkdir('.local',{recursive:true});await p.screenshot({path:`.local/plugin-usage-${width}-${mode}.png`});}await p.locator('#close-dialog').click();}}
});
