// Read existing local evidence only. Never start services or read credentials.
import fs from 'node:fs';
import path from 'node:path';
import crypto from 'node:crypto';

const args=process.argv.slice(2);
const validPrefix=value=>typeof value==='string'&&/^[a-z0-9]+(?:-[a-z0-9]+)*$/.test(value);
if(![2,4].includes(args.length) || args[0]!=='--evidence-prefix' || !validPrefix(args[1]) ||
   (args.length===4&&(args[2]!=='--ui-evidence-prefix'||!validPrefix(args[3])))) {
  console.error('Usage: node scripts/check-local-admin-delivery.mjs --evidence-prefix lowercase-name [--ui-evidence-prefix lowercase-name]');
  process.exit(2);
}
const prefix=args[1],uiPrefix=args[3]||prefix,reusedUIEvidence=uiPrefix!==prefix,failures=[],checks={};
function check(name,passed) {checks[name]=!!passed;if(!passed) failures.push(name);}
function read(file) {try{return fs.readFileSync(file,'utf8').replace(/^\uFEFF/,'');}catch{failures.push('无法读取 '+file);return null;}}
function json(file) {const value=read(file);if(value===null)return null;try{return JSON.parse(value);}catch{failures.push('无法解析 '+file);return null;}}
function hash(file) {try{return crypto.createHash('sha256').update(fs.readFileSync(file)).digest('hex').toUpperCase();}catch{failures.push('无法核对文件摘要 '+file);return null;}}
function modified(file) {try{return fs.statSync(file).mtimeMs;}catch{failures.push('无法读取修改时间 '+file);return NaN;}}
function counts(text) {
  const stats={};for(const name of ['tests','pass','fail','cancelled','skipped','todo'])stats[name]=Number(text?.match(new RegExp('(?:^|\\n)[^\\n\\w]*'+name+' (\\d+)\\s*(?:\\r?\\n|$)'))?.[1] ?? NaN);
  return stats;
}
const base='.local/'+prefix;
const uiBase='.local/'+uiPrefix,uiSuffixes=['-ui-tests.txt','-focused-ui.txt','-layout-result.json'];
const uiEvidenceFiles=uiSuffixes.map(suffix=>({file:uiPrefix+suffix,sha256:hash(uiBase+suffix)}));
const ui=counts(read(uiBase+'-ui-tests.txt')),focused=counts(read(uiBase+'-focused-ui.txt'));
for(const [name,stats] of [['界面回归',ui],['新增界面回归',focused]])check(name,Number.isSafeInteger(stats.tests)&&stats.tests>0&&stats.pass===stats.tests&&['fail','cancelled','skipped','todo'].every(key=>stats[key]===0));
const layout=json(uiBase+'-layout-result.json');
check('离线布局回归',layout?.passed===true&&layout.offline===true&&layout.networkRequestsAllowed===false&&Array.isArray(layout.measurements)&&layout.measurements.length>0);
const go=read(base+'-go-tests.txt');
for(const name of ['internal/localadmin','internal/basiccommands','plugin/llm','plugin/wordcount'])check('Go '+name,!!go&&new RegExp('^ok\\s+github\\.com/FloatTech/ZeroBot-Plugin/'+name+'\\s','m').test(go));
check('Go 无失败记录',!!go&&!/^FAIL\b/m.test(go));
const vet=read(base+'-vet.txt');check('vet 无诊断',vet!==null&&vet.trim()==='');
const build=read(base+'-build.txt');
const buildHasReadyMarker=!!build&&build.includes('Build ready:');
const buildSnapshot=build&&!buildHasReadyMarker?json(base+'-before.json'):null;
const smokeText=read(base+'-smoke.txt'),run=smokeText?.match(/ARTIFACTS=[^\r\n]*[\\/](run-\d+)\s*(?:\r?\n|$)/)?.[1];
check('整机产物路径',!!run);
const smoke=run?json('.local/smoke-admin/'+run+'/result.json'):null;
check('Windows 整机回归',smoke?.passed===true);
const smokeFlags=Object.entries(smoke||{}).filter(([key,value])=>key.endsWith('Passed')&&typeof value==='boolean');
check('整机检查标记',smokeFlags.length>0&&smokeFlags.every(([,value])=>value===true));
const deployment=json(base+'-deployment.json');
const required=['programMatchesStaged','startupConfigUnchanged','controlDatabaseUnchanged','dailyCredentialUnchanged','napcatProcessPreserved','qqLoggedIn','onebotReadOnlyAPIsPassed','singleZeroBotInstance','processIdentityVerified'];
for(const key of required)check('原部署 '+key,deployment?.[key]===true);
check('原部署后台及启动检查',deployment?.adminHTTPStatus===200&&deployment.startupErrorDetected===false);
const deploymentFlags=Object.entries(deployment||{}).filter(([key,value])=>typeof value==='boolean'&&key!=='startupErrorDetected');
check('原部署检查标记',deploymentFlags.length>0&&deploymentFlags.every(([,value])=>value===true));
const live='.local/bin/zerobot.exe',staged='.local/bin/zerobot-'+prefix+'-tested.exe';
const programSHA256=hash(live),testedSHA256=hash(staged);
check('构建就绪',!!build&&build.includes('all modules verified')&&(
  (buildHasReadyMarker&&build.includes('zerobot-'+prefix+'-tested.exe'))||
  (!buildHasReadyMarker&&/^go version go\d/m.test(build)&&typeof buildSnapshot?.stagedHash==='string'&&
    /^[A-F0-9]{64}$/.test(buildSnapshot.stagedHash)&&!!testedSHA256&&buildSnapshot.stagedHash===testedSHA256)));
check('磁盘程序与测试构建一致',!!programSHA256&&programSHA256===testedSHA256);
const stagedTime=modified(staged),smokeTime=Date.parse(smoke?.completedAt),deploymentTime=Date.parse(deployment?.completedAt);
check('构建整机部署时间顺序',Number.isFinite(stagedTime)&&Number.isFinite(smokeTime)&&Number.isFinite(deploymentTime)&&stagedTime<=smokeTime&&smokeTime<=deploymentTime);
check('界面记录先于部署',Number.isFinite(deploymentTime)&&uiSuffixes.every(suffix=>modified(uiBase+suffix)<=deploymentTime));

let linkCount=0,assetMarkerCount=0;const missingLinks=[],missingMarkers=[],assetHashes=[];
const docs='docs/local-admin',assets='internal/localadmin/assets';
try {
  for(const file of fs.readdirSync(docs).filter(name=>name.endsWith('.md'))) {
    const text=read(path.join(docs,file));if(text===null)continue;
    for(const match of text.matchAll(/\[[^\]]*\]\(([^)]+)\)/g)) {
      const target=match[1].split('#')[0];if(!target||/^(?:https?:|app:|codex:)/i.test(target))continue;
      linkCount++;if(!fs.existsSync(path.resolve(docs,target)))missingLinks.push({file,target});
    }
  }
  for(const file of fs.readdirSync(assets).filter(name=>/\.(?:html|js|css)$/.test(name))) {
    const target=path.join(assets,file),sha256=hash(target),mtime=modified(target);assetHashes.push({file,sha256});
    check('资源未晚于构建 '+file,Number.isFinite(stagedTime)&&mtime<=stagedTime);
  }
}catch{failures.push('文档或资源目录无法检查');}
const smokeSource=read('scripts/smoke-local-admin.go');
for(const match of smokeSource?.matchAll(/\{"\/(.*?)", "(?:text\/html|text\/javascript|text\/css); charset=utf-8", `([^`]+)`\}/g)||[]) {
  assetMarkerCount++;const file=match[1]||'index.html';
  if(!/^[A-Za-z0-9_.-]+$/.test(file)){missingMarkers.push({file});continue;}
  const text=read(path.join(assets,file));if(text===null||!text.includes(match[2]))missingMarkers.push({file,marker:match[2]});
}
check('文档链接',linkCount>0&&!missingLinks.length);
check('整机资源入口',assetMarkerCount>0&&!missingMarkers.length);
let uiBaseline=null;
if(reusedUIEvidence) {
  uiBaseline=json(uiBase+'-ui-baseline.json');
  check('界面基线来源',uiBaseline?.version===1&&uiBaseline.evidencePrefix===uiPrefix&&
    Number.isFinite(Date.parse(uiBaseline.checkedAt))&&Number.isFinite(Date.parse(uiBaseline.sourceTestTimes?.deployment))&&
    Date.parse(uiBaseline.sourceTestTimes.deployment)<=deploymentTime);
  const validHashes=entries=>Array.isArray(entries)&&entries.every(entry=>typeof entry?.file==='string'&&/^[A-Za-z0-9_.-]+$/.test(entry.file)&&typeof entry.sha256==='string'&&/^[A-F0-9]{64}$/.test(entry.sha256))&&new Set(entries.map(entry=>entry.file)).size===entries.length;
  const sameHashes=(expected,actual)=>validHashes(expected)&&expected.length===actual.length&&actual.every(entry=>entry.sha256&&expected.some(old=>old.file===entry.file&&old.sha256===entry.sha256));
  check('原界面证据摘要一致',sameHashes(uiBaseline?.uiEvidenceFiles,uiEvidenceFiles));
  check('全部界面资源保持一致',assetHashes.length>0&&sameHashes(uiBaseline?.assetHashes,assetHashes));
}
const result={checkedAt:new Date().toISOString(),passed:!failures.length,evidencePrefix:prefix,uiEvidencePrefix:uiPrefix,reusedUIEvidence,buildReadyMarkerRecorded:buildHasReadyMarker,buildHashFromUpgradeSnapshot:buildSnapshot?.stagedHash||null,testsExecuted:false,serviceAccessed:false,programSHA256,testedSHA256,sourceTestTimes:{smoke:smoke?.completedAt||null,deployment:deployment?.completedAt||null,uiBaseline:uiBaseline?.checkedAt||null,uiDeployment:uiBaseline?.sourceTestTimes?.deployment||deployment?.completedAt||null},ui,focused,layoutChecks:layout?.measurements?.length||0,smokeRun:run||null,linkCount,assetMarkerCount,checks,failures,missingLinks,missingMarkers,assetHashes,uiEvidenceFiles};
fs.mkdirSync('.local',{recursive:true});fs.writeFileSync('.local/current-delivery-evidence.json',JSON.stringify(result,null,2)+'\n');
if(result.passed&&!reusedUIEvidence) {
  fs.writeFileSync(uiBase+'-ui-baseline.json',JSON.stringify({version:1,evidencePrefix:uiPrefix,checkedAt:result.checkedAt,sourceTestTimes:result.sourceTestTimes,uiEvidenceFiles,assetHashes},null,2)+'\n');
}
console.log(JSON.stringify({passed:result.passed,evidencePrefix:prefix,uiEvidencePrefix:uiPrefix,reusedUIEvidence,uiTests:ui.tests,layoutChecks:result.layoutChecks,links:linkCount,assetMarkers:assetMarkerCount,testsExecuted:false,serviceAccessed:false,failures}));
process.exitCode=result.passed?0:1;
