// Appearance behavior and palette checks; no browser session or credentials.
import assert from 'node:assert/strict';
import fs from 'node:fs';
import vm from 'node:vm';
import test from 'node:test';

const key='zerobot.localadmin.theme.v1';
const script=fs.readFileSync(new URL('../internal/localadmin/assets/theme.js',import.meta.url),'utf8');
function harness({stored=null,dark=false,blocked=false,loading=true,noMedia=false}={}) {
  const events=new Map(),documentEvents=new Map(),writes=[],controls=[0,1].map(()=>({value:'',listener:null,addEventListener(name,listener){this.listener=listener;}}));
  let mediaListener;
  const media={matches:dark,addEventListener(name,listener){mediaListener=listener;}};
  const document={readyState:loading?'loading':'complete',documentElement:{dataset:{}},querySelectorAll:()=>loading?[]:controls,
    addEventListener:(name,listener)=>documentEvents.set(name,listener)};
  const window={localStorage:{getItem:()=>{if (blocked) throw new Error('blocked'); return stored;},setItem:(name,value)=>{if (blocked) throw new Error('blocked'); writes.push([name,value]); stored=value;}},
    matchMedia:()=>{if (noMedia) throw new Error('unsupported'); return media;},addEventListener:(name,listener)=>events.set(name,listener)};
  vm.runInNewContext(script,{window,document});
  return {controls,writes,root:document.documentElement,
    ready(){loading=false; documentEvents.get('DOMContentLoaded')?.();},
    select(value,index=0){controls[index].value=value; controls[index].listener?.();},
    system(dark){media.matches=dark; mediaListener?.();},
    storage(value,eventKey=key){stored=value; events.get('storage')?.({key:eventKey});}};
}
test('system follows OS before controls load; explicit choice survives reload and ignores OS changes',()=>{
  const h=harness({dark:true}); assert.equal(h.root.dataset.theme,'dark'); h.ready();
  assert.deepEqual(h.controls.map(c=>c.value),['system','system']);
  h.system(false); assert.equal(h.root.dataset.theme,'light');
  h.select('dark'); h.system(false); assert.equal(h.root.dataset.theme,'dark');
  assert.deepEqual(h.controls.map(c=>c.value),['dark','dark']);
  assert.deepEqual(h.writes,[[key,'dark']]);
  const reopened=harness({stored:'dark',dark:false}); assert.equal(reopened.root.dataset.theme,'dark');
  h.select('system',1); assert.equal(h.root.dataset.theme,'light'); h.system(true); assert.equal(h.root.dataset.theme,'dark');
});
test('other tabs and clearing storage update both controls without writing credentials or echoing invalid values',()=>{
  const h=harness({stored:'light',dark:true,loading:false});
  h.storage('dark'); assert.equal(h.root.dataset.theme,'dark');
  assert.deepEqual(h.controls.map(c=>c.value),['dark','dark']); assert.equal(h.writes.length,0);
  h.storage('<script>invalid</script>'); assert.equal(h.root.dataset.theme,'dark');
  assert.deepEqual(h.controls.map(c=>c.value),['system','system']);
  h.select('light'); h.storage(null,null); assert.equal(h.root.dataset.theme,'dark');
  h.storage('light','unrelated-account-setting'); assert.equal(h.root.dataset.theme,'dark');
  h.select('unknown'); assert.equal(h.root.dataset.theme,'dark'); assert.equal(h.writes.length,1);
});
test('blocked storage and unavailable matchMedia leave appearance controls usable',()=>{
  const h=harness({blocked:true,noMedia:true,loading:false}); assert.equal(h.root.dataset.theme,'light');
  h.select('dark'); assert.equal(h.root.dataset.theme,'dark');
  assert.deepEqual(h.controls.map(c=>c.value),['dark','dark']); assert.equal(h.writes.length,0);
});

const css=fs.readFileSync(new URL('../internal/localadmin/assets/appearance.css',import.meta.url),'utf8');
function palette(block) {return Object.fromEntries([...block.matchAll(/--([\w-]+):\s*(#[\da-f]{6})/gi)].map(match=>[match[1],match[2]]));}
function luminance(hex) {const rgb=[1,3,5].map(i=>parseInt(hex.slice(i,i+2),16)/255).map(c=>c<=.04045?c/12.92:((c+.055)/1.055)**2.4); return rgb[0]*.2126+rgb[1]*.7152+rgb[2]*.0722;}
function contrast(a,b) {const [lo,hi]=[luminance(a),luminance(b)].sort((a,b)=>a-b); return (hi+.05)/(lo+.05);}
test('both palettes keep normal, muted, action and warning/error text legible',()=>{
  const light=palette(css.match(/:root\s*\{([^}]+)\}/)[1]);
  const dark={...light,...palette(css.match(/:root\[data-theme="dark"\]\s*\{([^}]+)\}/)[1])};
  for (const colors of [light,dark]) {
    for (const [fg,bg] of [['text','page'],['text','surface'],['muted','surface'],['muted','surface-alt'],['accent','surface-alt'],['primary-text','green'],['primary-text','primary-hover'],['soft-text','soft'],['error-text','error-bg'],['warning-text','warning-bg']]) {
      assert.ok(contrast(colors[fg],colors[bg])>=4.5,`${fg}/${bg} has insufficient contrast`);
    }
  }
});
