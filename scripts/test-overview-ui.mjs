// Tests the real overview renderer/controller with an offline DOM and clock.
import assert from 'node:assert/strict';
import fs from 'node:fs';
import vm from 'node:vm';
import test from 'node:test';
class Element {
  constructor(){this.children=[];this.listeners=new Map();this.classes=new Set();this.value='';this.hidden=false;this._text='';this.dataset={};}
  get textContent(){return this._text;}
  set textContent(value){this._text=String(value);this.children=[];}
  set innerHTML(value){throw new Error('Overview must render text safely');}
  get classList(){return {add:name=>this.classes.add(name),toggle:(name,on)=>on?this.classes.add(name):this.classes.delete(name)};}
  append(...items){this.children.push(...items);}
  replaceChildren(...items){this.children=[...items];}
  // The overview fixture contains no plugin cards or their draft buttons.
  querySelectorAll(){return [];}
  addEventListener(name,callback){this.listeners.set(name,callback);}
  dispatch(name){return this.listeners.get(name)?.({preventDefault(){}});}
}
const app=fs.readFileSync(new URL('../internal/localadmin/assets/app.js',import.meta.url),'utf8');
const controller=fs.readFileSync(new URL('../internal/localadmin/assets/overview.js',import.meta.url),'utf8');
function harness(){
  const elements=new Map(),requests=[],timers=new Map();let timer=0;
  const get=id=>{if(!elements.has(id))elements.set(id,new Element());return elements.get(id);};
  const document=new Element();document.hidden=false;document.createElement=()=>new Element();
  const context=vm.createContext({document,$:get,authenticated:true,AbortController,duration:seconds=>String(seconds),
    setTimeout:(callback,delay)=>{timers.set(++timer,{callback,delay});return timer;},clearTimeout:id=>timers.delete(id),
    showError:(id,error)=>{get(id).textContent=error.message;get(id).hidden=false;},
    api:(path,body,signal)=>new Promise((resolve,reject)=>{const r={path,signal,resolve,reject};requests.push(r);signal.addEventListener('abort',()=>reject(Object.assign(new Error('aborted'),{name:'AbortError'})),{once:true});})});
  vm.runInContext(app.match(/function renderOverview\(value\) \{[\s\S]*?\n\}(?=\r?\nfunction renderPlugins)/)[0],context);
  vm.runInContext(controller,context);
  return {get,requests,timers,document,context,run:code=>vm.runInContext(code,context)};
}
const flush=()=>new Promise(resolve=>setImmediate(resolve));
const row=(message='test warning')=>({time:'2026-10-03T02:00:00Z',level:'warning',message});
const value=(alerts={warningCount:1,errorCount:0,entries:[row()],evicted:false})=>({botAccounts:['87654321'],pluginCount:108,uptimeSeconds:60,version:'fixture',goVersion:'fixture',configuredConnections:1,checkedAt:'2026-10-03T02:00:00Z',connectionState:'connected',alerts});
const modelStats=(patch={})=>{
  const zero={started:0,active:0,completed:0,failed:0,incomplete:0,averageMilliseconds:0,maxMilliseconds:0};
  return {startedAt:'2026-10-13T01:00:00Z',checkedAt:'2026-10-13T01:05:00Z',...zero,categories:['gpt','chat','agent','vision'].map(kind=>({kind,...zero})),...patch};
};

test('model statistics show zero requests and fixed categories without an inferred success',()=>{
  const h=harness();h.context.stats=modelStats();h.run('renderModelRequests(stats)');
  assert.match(h.get('model-request-summary').textContent,/尚无模型接口请求/);
  assert.equal(h.get('model-request-categories').children.length,4);
  assert.equal(h.get('model-request-categories').children[0].children[0].textContent,'GPT / 群总结');
  assert.match(h.get('model-request-categories').children[0].children[2].textContent,/耗时待统计/);
});
test('model statistics distinguish active completed failed and early closed exchanges and durations',()=>{
  const h=harness(),stats=modelStats({started:5,active:1,completed:2,failed:1,incomplete:1,averageMilliseconds:1250,maxMilliseconds:2000});
  stats.categories[0]={...stats.categories[0],started:5,active:1,completed:2,failed:1,incomplete:1,averageMilliseconds:1250,maxMilliseconds:2000};
  h.context.stats=stats;h.run('renderModelRequests(stats)');const text=h.get('model-request-summary').textContent;
  assert.match(text,/发起 5 次；进行中 1 次；接口完成 2 次；失败 1 次；提前关闭 1 次/);assert.match(text,/平均 1.250 秒，最长 2.000 秒/);
});
test('unavailable and malformed model statistics clear prior values without rendering untrusted category text',()=>{
  const h=harness();for(const malformed of [null,undefined,modelStats({started:-1}),modelStats({started:NaN}),modelStats({maxMilliseconds:Infinity}),modelStats({startedAt:'<script>bad</script>'}),modelStats({categories:[null,null,null,null]}),modelStats({categories:Array(4).fill({...modelStats().categories[0],kind:'<img onerror=bad>'})}),modelStats({started:1}),modelStats({averageMilliseconds:3,maxMilliseconds:2})]){
    h.context.stats=modelStats();h.run('renderModelRequests(stats)');h.context.stats=malformed;h.run('renderModelRequests(stats)');assert.equal(h.get('model-request-categories').children.length,0);assert.match(h.get('model-request-summary').textContent,/暂不可读取/);
  }
});
test('model statistics follow existing request cancellation and stale state clearing',async()=>{
  const h=harness();h.run('showOverview(true)');h.requests[0].resolve({...value(),modelRequests:modelStats()});await flush();assert.equal(h.get('model-request-categories').children.length,4);
  h.run('showOverview(false)');assert.equal(h.get('model-request-categories').children.length,0);assert.match(h.get('model-request-summary').textContent,/待更新/);
  h.run('showOverview(true)');h.requests[1].resolve({...value(),modelRequests:modelStats()});h.run('authenticated=false;resetOverview()');await flush();assert.equal(h.get('model-request-categories').children.length,0);
});
test('new process statistics replace old totals and a refresh error removes stale model data',async()=>{
  const h=harness();h.run('showOverview(true)');const old=modelStats({started:1,completed:1,averageMilliseconds:1,maxMilliseconds:1});old.categories[0]={...old.categories[0],started:1,completed:1,averageMilliseconds:1,maxMilliseconds:1};
  h.requests[0].resolve({...value(),modelRequests:old});await flush();assert.match(h.get('model-request-summary').textContent,/发起 1 次/);
  h.run('refreshOverview()');h.requests[1].resolve({...value(),modelRequests:modelStats({startedAt:'2026-10-13T02:00:00Z'})});await flush();assert.match(h.get('model-request-summary').textContent,/发起 0 次/);
  h.run('refreshOverview()');h.requests[2].reject(new Error('fixture'));await flush();assert.equal(h.get('model-request-categories').children.length,0);
});

test('overview renders literal alert text and distinguishes unavailable from no alerts',()=>{
  const h=harness(),payload='<img src=x onerror=alert(1)><script>bad</script>';
  h.context.value=value({warningCount:3,errorCount:2,entries:[row(payload)],evicted:true});h.run('renderOverview(value)');
  assert.equal(h.get('overview-alerts').children[0].children[1].textContent,payload);
  assert.match(h.get('overview-alert-summary').textContent,/警告 3 条，错误 2 条/);
  assert.match(h.get('overview-alert-summary').textContent,/已淘汰/);
  h.context.value=value(null);h.run('renderOverview(value)');
  assert.equal(h.get('overview-alerts').children.length,0);assert.match(h.get('overview-alert-summary').textContent,/暂不可读取/);
  assert.equal(h.get('connection-badge').textContent,'已连接');assert.equal(h.get('accounts').children.length,1);
  h.context.value=value({warningCount:0,errorCount:0,entries:[],evicted:false});h.run('renderOverview(value)');
  assert.match(h.get('overview-alerts').children[0].textContent,/暂无警告或错误/);
  h.context.value={...value(),connectionState:'waiting',botAccounts:[]};h.run('renderOverview(value)');
  assert.equal(h.get('connection-badge').textContent,'等待连接');assert.equal(h.get('account-count').textContent,'0');
});
test('polling cancels on navigation, hiding and logout and refuses stale completion',async()=>{
  const h=harness();h.run('showOverview(true)');assert.equal(h.requests.length,1);
  h.requests[0].resolve(value());h.run('showOverview(false)');await flush();
  assert.equal(h.get('accounts').children.length,0);assert.equal(h.timers.size,0);
  h.run('showOverview(true)');h.document.hidden=true;h.document.dispatch('visibilitychange');
  assert.equal(h.requests[1].signal.aborted,true);await flush();assert.equal(h.timers.size,0);
  h.document.hidden=false;h.document.dispatch('visibilitychange');assert.equal(h.requests.length,3);
  h.requests[2].resolve(value());h.run('authenticated=false;resetOverview()');await flush();
  assert.equal(h.get('overview-alerts').children.length,0);assert.equal(h.get('accounts').children.length,0);assert.equal(h.timers.size,0);
  h.run('authenticated=true;showOverview(true)');h.requests[3].resolve(value());await flush();
  assert.equal(h.get('account-count').textContent,'1');assert.equal([...h.timers.values()][0].delay,10000);
});
test('manual and timed refresh do not overlap; failures clear old online status and recover',async()=>{
  const h=harness();h.run('showOverview(true);refreshOverview();refreshOverview()');assert.equal(h.requests.length,1);
  h.requests[0].resolve(value());await flush();assert.equal(h.get('connection-badge').textContent,'已连接');
  const tick=[...h.timers.values()][0];h.timers.clear();tick.callback();h.run('refreshOverview()');assert.equal(h.requests.length,2);
  h.requests[1].reject(new Error('offline fixture'));await flush();
  assert.equal(h.get('connection-badge').textContent,'无法获取状态');assert.equal(h.get('accounts').children.length,0);
  assert.equal(h.get('checked-at').textContent,'—');assert.equal(h.get('account-count').textContent,'—');assert.equal(h.get('page-error').hidden,false);
  assert.equal(h.timers.size,1);h.run('refreshOverview()');assert.equal(h.timers.size,0);
  h.requests[2].resolve(value());await flush();assert.equal(h.get('page-error').hidden,true);assert.equal(h.get('connection-badge').textContent,'已连接');
  h.run('showOverview(false)');assert.equal(h.timers.size,0);
  assert.equal(h.get('page-error').hidden,true);
});

test('an aborted response body cannot expire a newer authenticated UI session',async()=>{
  const h=harness();h.document.getElementById=h.get;h.document.querySelectorAll=()=>[];
  let resolveBody;const body=new Promise(resolve=>{resolveBody=resolve;});
  const c=vm.createContext({document:h.document,window:new Element(),AbortController,fetch:async()=>({ok:false,status:401,json:()=>body})});
  vm.runInContext(app,c);
  vm.runInContext('authenticated=true;const oldController=new AbortController();var pending=api("/api/overview",undefined,oldController.signal).catch(error=>error.name);',c);
  await flush();vm.runInContext('oldController.abort()',c);resolveBody({error:'expired old session'});
  assert.equal(await c.pending,'AbortError');assert.equal(vm.runInContext('authenticated',c),true);
});
test('full app login, navigation and logout keep plugin failures separate from connection health',async()=>{
  const h=harness(),nav=new Element();nav.dataset.view='overview';
  const document=h.document;document.getElementById=h.get;document.querySelector=()=>nav;document.querySelectorAll=selector=>selector==='[data-view]'?[nav]:[];
  const fetches=[];
  h.context.fetch=(path,options)=>new Promise(resolve=>fetches.push({path,options,resolve}));
  // Evaluate app first in a new context, then the overview controller, as in HTML.
  const c=vm.createContext({document,window:Object.assign(new Element(),{getSelection:()=>null}),AbortController,$:h.get,fetch:h.context.fetch,
    setTimeout:h.context.setTimeout,clearTimeout:h.context.clearTimeout});
  vm.runInContext(app,c);vm.runInContext(controller,c);
  document.dispatch('DOMContentLoaded');
  fetches[0].resolve({ok:true,json:async()=>({authenticated:true,csrfToken:'fixture'})});await flush();
  const overview=fetches.find(r=>r.path==='/api/overview'),plugins=fetches.find(r=>r.path.startsWith('/api/plugins'));
  assert.ok(overview&&plugins);
  plugins.resolve({ok:false,status:500,json:async()=>({error:'plugin fixture failed'})});
  overview.resolve({ok:true,json:async()=>value()});await flush();
  assert.equal(h.get('connection-badge').textContent,'已连接');assert.equal(h.get('page-error').hidden,true);
  assert.equal(h.get('plugin-error').textContent,'plugin fixture failed');
  nav.dataset.view='plugins';nav.dispatch('click');assert.equal(h.get('connection-badge').textContent,'状态待更新');
  assert.equal(h.timers.size,0);
  h.get('logout').dispatch('click');await flush();const logout=fetches.find(r=>r.path==='/api/logout');assert.ok(logout);logout.resolve({ok:true,json:async()=>({})});await flush();
  assert.equal(h.get('workspace').hidden,true);assert.equal(h.get('auth').hidden,false);assert.equal(h.timers.size,0);
});
