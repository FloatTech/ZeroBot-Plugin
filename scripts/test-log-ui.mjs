// Offline DOM harness: exercises the log page without browser/site permissions.
import assert from 'node:assert/strict';
import fs from 'node:fs';
import vm from 'node:vm';
import test from 'node:test';

class Element {
  constructor(tag='div') { this.tagName=tag; this.children=[]; this.dataset={}; this.listeners=new Map(); this.value=''; this.className=''; this.hidden=false; this.checked=false; this.scrollTop=0; this._text=''; }
  get textContent() { return this._text; }
  set textContent(value) { this._text=String(value); this.replaceChildren(); }
  set innerHTML(value) { throw new Error('Log text must never enter innerHTML'); }
  get classList() { return {add: name=>{this.className+=' '+name;}}; }
  get firstElementChild() { return this.children[0]; }
  get scrollHeight() { return this.children.length*40; }
  append(...items) { for (const item of items) {item.parent=this; this.children.push(item);} }
  replaceChildren(...items) { for (const item of this.children) item.parent=null; this.children=[]; this.append(...items); }
  remove() { if (this.parent) {this.parent.children=this.parent.children.filter(item=>item!==this); this.parent=null;} }
  querySelector(selector) { return this.children.find(item=>selector==='.empty' && item.className==='empty') || null; }
  querySelectorAll(selector) { return this.children.flatMap(item=>[...(selector==='.copy-log-entry' && item.className.split(' ').includes('copy-log-entry')?[item]:[]),...item.querySelectorAll(selector)]); }
  setAttribute(name,value) { this[name]=String(value); }
  addEventListener(name,callback) {this.listeners.set(name,callback);}
  dispatch(name) {this.listeners.get(name)?.({preventDefault(){}});}
}

function harness() {
  const elements=new Map(), timers=new Map(), requests=[], exports=[], downloads=[], revoked=[];
  const document=new Element(); document.hidden=false; document.createElement=tag=>new Element(tag);
  document.body=new Element('body');
  document.createElement=tag=>{const item=new Element(tag); item.click=()=>downloads.push({href:item.href,filename:item.download}); return item;};
  const get=id=>{if (!elements.has(id)) elements.set(id,new Element()); return elements.get(id);};
  const window=new Element(); let timerID=0;
  const context=vm.createContext({document,window,$:get,authenticated:true,csrf:'fixture',AbortController,URLSearchParams,
    URL:{createObjectURL:()=> 'blob:fixture-'+downloads.length,revokeObjectURL:url=>revoked.push(url)},
    fetch:(path,options)=>new Promise((resolve,reject)=>{exports.push({path,options,resolve,reject}); options.signal.addEventListener('abort',()=>reject(Object.assign(new Error('aborted'),{name:'AbortError'})),{once:true});}),
    showAuth:()=>vm.runInContext('resetLogs()',context),
    setTimeout:callback=>{timers.set(++timerID,callback); return timerID;}, clearTimeout:id=>timers.delete(id),
    showError:(id,error)=>{get(id).textContent=error.message; get(id).hidden=false;},
    api:(path,body,signal)=>new Promise((resolve,reject)=>{
      const request={path,signal,resolve,reject}; requests.push(request);
      signal.addEventListener('abort',()=>reject(Object.assign(new Error('aborted'),{name:'AbortError'})),{once:true});
    })});
  get('log-autoscroll').checked=true;
  vm.runInContext(fs.readFileSync(new URL('../internal/localadmin/assets/logs.js',import.meta.url),'utf8'),context);
  return {get,context,timers,requests,exports,downloads,revoked,document,run:code=>vm.runInContext(code,context)};
}
const flush=()=>new Promise(resolve=>setImmediate(resolve));
const row=(id,text='ordinary log')=>({id:String(id),time:'2026-10-02T12:00:00Z',level:'info',message:text});
const page=(entries,cursor=entries.at(-1)?.id || '0')=>({entries,cursor,instance:'0123456789abcdef',retained:entries.length,capacity:1000,gap:false,reset:false});

test('log page renders literal text, deduplicates, bounds DOM and controls scrolling',()=>{
  const h=harness(), payload='<img src=x onerror=alert(1)><script>bad()</script>';
  h.context.value=page([row(1,payload)]);
  h.run('renderLogs(value); renderLogs(value)');
  const container=h.get('log-records');
  assert.equal(container.children.length,1);
  assert.equal(container.firstElementChild.children[1].textContent,payload);
  assert.equal(container.scrollTop,40);
  h.get('log-autoscroll').checked=false; container.scrollTop=13;
  h.context.value=page(Array.from({length:1001},(_,index)=>row(index+2)));
  h.run('renderLogs(value)');
  assert.equal(container.children.length,1000);
  assert.equal(h.run('logSeen.size'),1000);
  assert.equal(container.scrollTop,13);
  h.context.value={...page([row(1,'new process')]),reset:true};
  h.run('renderLogs(value)');
  assert.equal(container.children.length,1);
  assert.equal(h.get('log-gap').hidden,false);
});

test('navigation, pause, hidden page and logout cancel polling without stale render',async()=>{
  const h=harness(); h.run('showLogs(true)');
  assert.equal(h.requests.length,1);
  h.run('showLogs(false)');
  assert.equal(h.requests[0].signal.aborted,true);
  await flush(); assert.equal(h.timers.size,0);
  h.run('showLogs(true)');
  h.requests[1].resolve(page([row(1)])); await flush(); assert.equal(h.timers.size,1);
  h.get('pause-logs').dispatch('click');
  assert.equal(h.timers.size,0); assert.equal(h.run('logPaused'),true);
  h.get('pause-logs').dispatch('click');
  assert.equal(h.requests.length,3);
  h.document.hidden=true; h.document.dispatch('visibilitychange');
  assert.equal(h.requests[2].signal.aborted,true); await flush(); assert.equal(h.timers.size,0);
  h.document.hidden=false; h.document.dispatch('visibilitychange');
  assert.equal(h.requests.length,4);
  // A resolved old request can still resume its async continuation after logout.
  h.requests[3].resolve(page([row(99,'stale after logout')]));
  h.run('authenticated=false; resetLogs()'); await flush();
  assert.equal(h.get('log-records').children.length,0); assert.equal(h.timers.size,0);
});

test('filters restart the cursor and errors retry only while log view remains active',async()=>{
  const h=harness(); h.run('showLogs(true)');
  h.get('log-level').value='warning'; h.get('log-keyword').value='测试 keyword';
  h.get('log-filter').dispatch('submit');
  assert.equal(h.requests[0].signal.aborted,true);
  const query=new URL(h.requests[1].path,'http://fixture.invalid').searchParams;
  assert.equal(query.get('after'),'0'); assert.equal(query.get('q'),'测试 keyword'); assert.equal(query.get('level'),'warning');
  h.requests[1].reject(new Error('synthetic failure')); await flush();
  assert.equal(h.get('log-error').hidden,false); assert.equal(h.timers.size,1);
  h.run('showLogs(false)'); assert.equal(h.timers.size,0);
});

const exported=(overrides={})=>({ok:true,status:200,
  headers:new Headers({'Content-Type':'text/plain; charset=utf-8','Content-Length':'24','Content-Disposition':'attachment; filename="zerobot-logs-20261002T120000Z.txt"'}),
  blob:async()=>new Blob(['脱敏日志\n']),...overrides});

test('export uses applied filters, works while paused, prevents duplicates and releases download URL',async()=>{
  const h=harness(); h.run('showLogs(true)');
  h.get('log-level').value='warning'; h.get('log-keyword').value='  测试  ';
  h.get('log-filter').dispatch('submit'); h.get('pause-logs').dispatch('click');
  h.get('log-keyword').value='unapplied draft';
  h.get('export-logs').dispatch('click'); h.get('export-logs').dispatch('click');
  assert.equal(h.exports.length,1); assert.equal(h.get('export-logs').disabled,true);
  const query=new URL(h.exports[0].path,'http://fixture.invalid').searchParams;
  assert.equal(query.get('level'),'warning'); assert.equal(query.get('q'),'测试'); assert.equal(query.has('after'),false);
  assert.equal(h.exports[0].options.credentials,'same-origin');
  h.exports[0].resolve(exported()); await flush();
  assert.equal(h.downloads.length,1); assert.equal(h.downloads[0].filename,'zerobot-logs-20261002T120000Z.txt');
  assert.equal(h.document.body.children.length,0); assert.equal(h.get('export-logs').disabled,false);
  for (const callback of h.timers.values()) callback();
  assert.deepEqual(h.revoked,[h.downloads[0].href]);
});

test('export cancellation and stale body completion never download after navigation, filters, hiding or logout',async()=>{
  for (const action of ['showLogs(false)','resetLogs()','reloadLogs()', 'document.hidden=true; document.dispatch("visibilitychange")']) {
    const h=harness(); h.run('showLogs(true)'); h.get('export-logs').dispatch('click');
    let finish; h.exports[0].resolve(exported({blob:()=>new Promise(resolve=>{finish=resolve;})})); await flush();
    h.run(action); finish(new Blob(['stale sensitive data'])); await flush();
    assert.equal(h.exports[0].options.signal.aborted,true); assert.equal(h.downloads.length,0); assert.equal(h.get('export-logs').disabled,false);
  }
});

test('failed, expired or malformed exports produce no file',async()=>{
  for (const response of [
    exported({ok:false,status:503,json:async()=>({error:'日志暂不可读取'})}),
    exported({ok:false,status:401,json:async()=>({error:'请登录'})}),
    exported({headers:new Headers({'Content-Type':'text/html','Content-Length':'24','Content-Disposition':'attachment; filename="unsafe.html"'})}),
    exported({headers:new Headers({'Content-Type':'text/plain','Content-Length':String(4*1024*1024),'Content-Disposition':'attachment; filename="zerobot-logs-20261002T120000Z.txt"'})}),
    exported({blob:async()=>({size:4*1024*1024})})
  ]) {
    const h=harness(); h.run('showLogs(true)'); h.get('export-logs').dispatch('click'); h.exports[0].resolve(response); await flush();
    assert.equal(h.downloads.length,0); assert.equal(h.get('export-logs').disabled,false);
    if (response.status===401) {assert.equal(h.context.authenticated,false); assert.equal(h.context.csrf,'');}
    else assert.notEqual(h.get('log-export-status').textContent,'');
  }
});

test('applied filter status distinguishes drafts; quick issue/export/clear use committed conditions and preserve pause',async()=>{
  const h=harness();h.run('showLogs(true)');
  h.get('log-keyword').value='<img src=x onerror=bad()>';h.get('log-keyword').dispatch('input');
  assert.match(h.get('log-filter-status').textContent,/尚未应用/);
  assert.doesNotMatch(h.get('log-filter-status').textContent,/<img/);
  h.get('log-filter').dispatch('submit');
  assert.ok(h.get('log-filter-status').textContent.includes('<img src=x onerror=bad()>'));
  h.get('pause-logs').dispatch('click');h.get('log-autoscroll').checked=false;
  h.get('log-keyword').value='unapplied';h.get('log-level').value='info';h.get('log-keyword').dispatch('input');
  h.get('only-issue-logs').dispatch('click');
  assert.equal(h.run('logLevel'),'issues');assert.equal(h.run('logPaused'),true);
  assert.equal(h.get('log-keyword').value,'<img src=x onerror=bad()>');assert.equal(h.get('log-autoscroll').checked,false);
  assert.doesNotMatch(h.get('log-filter-status').textContent,/尚未应用/);
  assert.equal(h.requests.length,2);h.get('export-logs').dispatch('click');
  const query=new URL(h.exports[0].path,'http://fixture.invalid').searchParams;
  assert.equal(query.get('level'),'issues');assert.equal(query.get('q'),'<img src=x onerror=bad()>');
  h.get('clear-log-filter').dispatch('click');await flush();
  assert.equal(h.exports[0].options.signal.aborted,true);assert.equal(h.get('log-level').value,'');assert.equal(h.get('log-keyword').value,'');
  assert.equal(h.run('logPaused'),true);assert.equal(h.requests.length,2);
  h.get('pause-logs').dispatch('click');
  const cleared=new URL(h.requests[2].path,'http://fixture.invalid').searchParams;
  assert.equal(cleared.get('level'),'');assert.equal(cleared.get('q'),'');assert.equal(cleared.get('after'),'0');
});

test('quick filter rejects stale resolved rows and logout resets applied/draft conditions',async()=>{
  const h=harness();h.run('showLogs(true)');
  h.requests[0].resolve(page([row(1,'old data')]));h.get('only-issue-logs').dispatch('click');await flush();
  assert.equal(h.get('log-records').children.length,0);assert.equal(h.requests[0].signal.aborted,true);
  h.requests[1].resolve(page([{...row(2,'issue'),level:'warning'}]));await flush();
  assert.equal(h.get('log-records').children.length,1);
  h.get('log-keyword').value='private draft';h.run('authenticated=false;resetLogs()');
  assert.equal(h.run('logLevel'),'');assert.equal(h.run('logKeyword'),'');assert.equal(h.get('log-keyword').value,'');
  assert.doesNotMatch(h.get('log-filter-status').textContent,/尚未应用|private draft/);
  const count=h.requests.length;h.get('only-issue-logs').dispatch('click');h.get('clear-log-filter').dispatch('click');
  assert.equal(h.requests.length,count);assert.equal(h.get('log-records').children.length,0);
});
