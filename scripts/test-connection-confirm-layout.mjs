// Offline rendering of the actual confirmation markup and theme; no live browser session.
import fs from 'node:fs/promises';
import path from 'node:path';
import assert from 'node:assert/strict';
import {createRequire} from 'node:module';
const require=createRequire(import.meta.url);
const modules=process.env.LOCALADMIN_NODE_MODULES||path.join(process.env.USERPROFILE,'.cache/codex-runtimes/codex-primary-runtime/dependencies/node/node_modules');
const {chromium}=require(path.join(modules,'playwright'));
const html=(await fs.readFile('internal/localadmin/assets/index.html','utf8')).replace(/<script\b[^>]*>[\s\S]*?<\/script>/g,'').replace(/<link\b[^>]*>/g,'');
const css=await fs.readFile('internal/localadmin/assets/style.css','utf8')+'\n'+await fs.readFile('internal/localadmin/assets/appearance.css','utf8');
const out=process.env.CONNECTION_CONFIRM_LAYOUT_OUT||'.local/connection-confirm-layout';await fs.mkdir(out,{recursive:true});
const browser=await chromium.launch({channel:'chrome',headless:true});
const measurements=[];let passed=false;
try{
 const context=await browser.newContext();await context.route('**/*',r=>r.abort());const page=await context.newPage();await page.setContent(html);await page.addStyleTag({content:css});
 for(const kind of ['add','draft'])for(const width of [1280,768,375,320])for(const mode of ['light','dark']){
  await page.setViewportSize({width,height:900});
  await page.evaluate(({mode,kind})=>{document.documentElement.dataset.theme=mode;document.getElementById(`connection-${kind}-confirm-message`).textContent=kind==='add'?'新增本机反向监听并备份当前配置？保存不自动连接或监听，接入程序请使用相同地址与 Token。':'连接设置有未保存更改。切换连接将丢弃当前连接的地址与 Token 选择（主人账号草稿保留），是否继续？';document.getElementById(`connection-${kind}-confirmation`).showModal();},{mode,kind});
  const m=await page.evaluate(kind=>{const d=document.getElementById(`connection-${kind}-confirmation`),r=d.getBoundingClientRect();return {left:r.left,right:r.right,top:r.top,bottom:r.bottom,scrollWidth:d.scrollWidth,clientWidth:d.clientWidth,background:getComputedStyle(d).backgroundColor,surface:getComputedStyle(document.documentElement).getPropertyValue('--surface').trim(),focus:document.activeElement.id,buttons:[...d.querySelectorAll('button')].map(b=>{const r=b.getBoundingClientRect();return {left:r.left,right:r.right,top:r.top,bottom:r.bottom};})};},kind);
  measurements.push({kind,width,mode,...m});assert.ok(m.left>=0&&m.right<=width&&m.top>=0&&m.bottom<=900&&m.scrollWidth<=m.clientWidth+1);assert.equal(m.focus,`connection-${kind}-confirm-cancel`);for(const b of m.buttons)assert.ok(b.left>=m.left&&b.right<=m.right&&b.top>=m.top&&b.bottom<=m.bottom);
  await page.screenshot({path:path.join(out,`${kind}-${width}-${mode}.png`)});await page.evaluate(kind=>document.getElementById(`connection-${kind}-confirmation`).close(),kind);
 }
 passed=true;
}finally{await browser.close();await fs.writeFile(out+'/result.json',JSON.stringify({passed,offline:true,networkRequestsAllowed:false,measurements},null,2));}
console.log(JSON.stringify({passed,checks:measurements.length,out}));
