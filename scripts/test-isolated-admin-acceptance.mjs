// Local CLI/API integration only. Synthetic isolated instances; no browser or live QQ.
import assert from 'node:assert/strict';
import fs from 'node:fs/promises';
import path from 'node:path';
import {spawn} from 'node:child_process';
import crypto from 'node:crypto';

const root=await fs.realpath(process.cwd());
const executable=path.join(root,'.local/bin/start-admin-acceptance-tested.exe');
const outputRoot=path.join(root,'.local/admin-acceptance');
const active=new Set(),profiles=[];
const readJSON=async file=>JSON.parse((await fs.readFile(file,'utf8')).replace(/^\uFEFF/,''));
function launch(args){
  const child=spawn(executable,args,{cwd:root,windowsHide:true,stdio:['pipe','pipe','pipe']});active.add(child);
  const messages=[],waiters=[];let buffer='',artifact='',errorText='';
  child.stdout.on('data',data=>{buffer+=data.toString();let end;while((end=buffer.indexOf('\n'))>=0){const line=buffer.slice(0,end).trim();buffer=buffer.slice(end+1);if(line.startsWith('ARTIFACTS='))artifact=line.slice(10);if(line.startsWith('{')){const value=JSON.parse(line);if(waiters.length)waiters.shift()(value);else messages.push(value);}}});
  child.stderr.on('data',data=>{errorText+=data.toString();});child.stdin.on('error',()=>{});
  const exited=new Promise((resolve,reject)=>{child.once('error',reject);child.once('close',(code,signal)=>{active.delete(child);resolve({code,signal});});});
  async function ready(){const value=messages.length?messages.shift():await deadline(new Promise(resolve=>waiters.push(resolve)),90000,'fixture readiness timed out');assert.equal(value.state,'ready');assert.match(value.url,/^http:\/\/127\.0\.0\.1:\d+\/$/);assert.notEqual(new URL(value.url).port,'3010');assert.ok(value.output.startsWith(outputRoot+path.sep));assert.equal(value.manualAcceptancePassed,false);return value;}
  async function result(expectedCode=0){const ended=await deadline(exited,150000,'fixture did not stop');assert.equal(ended.code,expectedCode,'unexpected fixture exit (inspect its local stderr record)');assert.ok(artifact.startsWith(outputRoot+path.sep));const value=await readJSON(path.join(artifact,'result.json'));assert.equal(value.instanceStopped,true);assert.equal(value.manualAcceptancePassed,false);assert.equal(value.realQQMessagesSent,false);assert.equal((await readJSON(path.join(artifact,'session.json'))).state,'stopped');return {artifact,...value};}
  return {child,ready,result,exited,get errorText(){return errorText;}};
}
async function deadline(promise,ms,message){let timer;try{return await Promise.race([promise,new Promise((_,reject)=>{timer=setTimeout(()=>reject(Error(message)),ms);})]);}finally{clearTimeout(timer);}}
async function request(url,endpoint,body,session={}){
  const base=new URL(url).origin;
  const response=await fetch(base+endpoint,{method:body===undefined?'GET':'POST',headers:{Origin:base,'Content-Type':'application/json','X-Local-Request':'1',...(session.cookie?{Cookie:session.cookie}:{}),...(session.csrf?{'X-CSRF-Token':session.csrf}:{})},...(body===undefined?{}:{body:JSON.stringify(body)}),signal:AbortSignal.timeout(10000)});
  const value=await response.json();return {status:response.status,value,cookie:response.headers.get('set-cookie')?.split(';')[0],csrf:value.csrfToken};
}
let passed=false;
try{
  for(const args of [['-binary-name','../zerobot.exe'],['-lifetime','1s']]){
    const fixture=launch(args);const exit=await deadline(fixture.exited,10000,'invalid arguments did not exit');assert.equal(exit.code,2);assert.match(fixture.errorText,/Invalid arguments/);
  }
  profiles.push({name:'invalid-arguments',passed:true});

  const interactive=launch(['-binary-name','zerobot.exe','-lifetime','2m']);
  const first=await interactive.ready();const password=(await fs.readFile(first.loginFile,'utf8')).trim().split(/\r?\n/).at(-1);
  const session=await request(first.url,'/api/login',{password});assert.equal(session.status,200);
  const changedPassword='QA-'+crypto.randomBytes(24).toString('hex');
  const change=await request(first.url,'/api/password',{currentPassword:password,newPassword:changedPassword,confirmPassword:changedPassword},session);assert.equal(change.status,200);assert.equal(change.value.passwordChanged,true);
  interactive.child.stdin.write('restart\n');const restarted=await interactive.ready();assert.notEqual(restarted.url,first.url);
  const oldSession=await request(restarted.url,'/api/overview',undefined,session);assert.equal(oldSession.status,401);
  const replacement=await request(restarted.url,'/api/login',{password:changedPassword});assert.equal(replacement.status,200);
  assert.ok((await fs.readFile(first.loginFile,'utf8'))===(await fs.readFile(restarted.loginFile,'utf8')),'fixture login file changed unexpectedly');
  interactive.child.stdin.write('status\n');assert.equal((await interactive.ready()).url,restarted.url);
  interactive.child.stdin.write('stop\n');const stopped=await interactive.result();assert.equal(stopped.stopReason,'stop-command');assert.equal(stopped.processStarts,2);assert.equal(stopped.environmentPrepared,true);
  profiles.push({name:'changed-password-restart-status-stop',passed:true,artifact:stopped.artifact});

  const eof=launch(['-binary-name','zerobot.exe','-lifetime','2m']);eof.child.stdin.end();await eof.ready();const eofResult=await eof.result();assert.equal(eofResult.stopReason,'input-ended');assert.equal(eofResult.processStarts,1);
  profiles.push({name:'input-ended',passed:true,artifact:eofResult.artifact});

  const timed=launch(['-binary-name','zerobot.exe','-lifetime','1m']);await timed.ready();const timedResult=await timed.result();assert.equal(timedResult.stopReason,'signal-or-time-limit');assert.equal(timedResult.processStarts,1);
  profiles.push({name:'expiry',passed:true,artifact:timedResult.artifact});
  passed=true;
}finally{
  for(const child of active){child.stdin.write('stop\n');child.stdin.end();}
  // Fixture lifetimes remain bounded even if a failing assertion ends this runner.
  if(active.size)await new Promise(resolve=>setTimeout(resolve,2000));
  const result={completedAt:new Date().toISOString(),passed,isolated:true,browserUsed:false,realQQMessagesSent:false,manualAcceptancePassed:false,profiles};
  await fs.writeFile('.local/isolated-acceptance-interactive-result.json',JSON.stringify(result,null,2)+'\n');
  console.log(JSON.stringify({passed,profiles:profiles.length,realQQMessagesSent:false,manualAcceptancePassed:false}));
}
