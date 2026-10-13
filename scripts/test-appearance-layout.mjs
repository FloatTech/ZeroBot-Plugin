// Offline rendering only: no local admin session, website navigation or network.
import fs from 'node:fs/promises';
import path from 'node:path';
import {createRequire} from 'node:module';
const require=createRequire(import.meta.url);
const modules=process.env.LOCALADMIN_NODE_MODULES || path.join(process.env.USERPROFILE,'.cache/codex-runtimes/codex-primary-runtime/dependencies/node/node_modules');
const {chromium}=require(path.join(modules,'playwright'));
const root=process.cwd(),out=path.join(root,'.local','appearance-layout');
await fs.mkdir(out,{recursive:true});
const html=(await fs.readFile('internal/localadmin/assets/index.html','utf8')).replace(/<script\b[^>]*>[\s\S]*?<\/script>/g,'').replace(/<link\b[^>]*>/g,'');
const css=await fs.readFile('internal/localadmin/assets/style.css','utf8')+'\n'+await fs.readFile('internal/localadmin/assets/appearance.css','utf8');
const theme=await fs.readFile('internal/localadmin/assets/theme.js','utf8');
const browser=await chromium.launch({channel:'chrome',headless:true});
let passed=false;
const measurements=[];
try {
  const context=await browser.newContext();
  await context.route('**/*',route=>route.abort());
  const page=await context.newPage();
  await page.setContent(html);
  await page.addStyleTag({content:css});
  await page.addScriptTag({content:theme});
  // Populate synthetic examples to exercise long IDs, URLs, plugin cards and forms.
  await page.evaluate(()=>{
    for (const [id,value] of Object.entries({'account-count':'1','plugin-count':'108','uptime':'2 小时 15 分','nav-count':'108','connection-detail':'后台运行正常，已接入模拟 OneBot。','connection-badge':'已连接','version':'测试构建','go-version':'go1.26.1','configured-count':'1','checked-at':'00:00:00'})) document.getElementById(id).textContent=value;
    document.getElementById('connection-url').value='ws://127.0.0.1:6700/a-long-endpoint-for-layout-testing';
    document.getElementById('overview-alert-summary').textContent='本次启动仍保留的日志：警告 3 条，错误 1 条。显示最新 5 条；历史记录不代表问题仍未解决。';
    document.getElementById('model-request-summary').textContent='统计起点：2026/10/13 09:00:00。发起 9007199254740000 次；进行中 10000 次；接口完成 8000000000000000 次；失败 900000 次；提前关闭 100000 次。已结束请求平均 12345.678 秒，最长 123456789.123 秒。';
    document.getElementById('model-request-categories').innerHTML=['GPT / 群总结','普通聊天','Agent','识图'].map(name=>'<article><h3>'+name+'</h3><p>发起 9007199254740000 次；进行中 10000 次；接口完成 8000000000000000 次；失败 900000 次；提前关闭 100000 次。</p><p class="field-help">已结束请求平均 12345.678 秒，最长 123456789.123 秒。</p></article>').join('');
    document.getElementById('overview-alerts').innerHTML='<article class="log-entry"><div class="log-meta"><time>2026/10/3 00:00:00</time><span class="log-level log-warning-level">WARNING</span></div><pre>用于布局检查的合成警告，synthetic-long-warning-for-wrap-testing-012345678901234567890123456789。</pre></article>';
    document.getElementById('connection-owners').value='12345678, 23456789';
    document.getElementById('startup-differences').hidden=false;document.getElementById('startup-difference-status').textContent='已保存配置与本次启动不同，待重启；6 个类别有变化。';
    document.getElementById('startup-difference-list').innerHTML=['正向 OneBot 连接','反向 OneBot 连接','机器人主人','其他机器人参数','其他启动设置','字段名称或表示差异'].map((label,i)=>'<article class="startup-difference-row"><strong>'+label+'</strong><span class="startup-difference-state difference-changed">有变化</span>'+(i<5?'<p class="field-help">本次启动 100000'+(i<2?'条连接':i===2?'个账号':'项设置')+' → 已保存 100000'+(i<2?'条连接':i===2?'个账号':'项设置')+'（数量相同，内容不同）</p>':'')+'</article>').join('');
    document.getElementById('connection-delete-preview').hidden=false;document.getElementById('connection-delete-target').textContent='删除目标：正向连接 2';document.getElementById('connection-delete-address').textContent='已保存地址：ws://127.0.0.1:6702/a-long-endpoint-for-deletion-layout-testing';document.getElementById('connection-delete-impact').textContent='已保存连接共 64 条，删除后保留 63 条。同类型序号会调整，本次启动信息仍按原序号展示。';
    document.getElementById('connection-draft-status').textContent='未保存更改：当前连接地址或 Token、主人账号。保存前请核对；切换后台页面保留草稿。';
    document.getElementById('connection-fields').disabled=false;document.getElementById('connection-add-fields').disabled=false;document.getElementById('connection-add-url').value='ws://127.0.0.1:6702';document.getElementById('connection-add-status').textContent='新增草稿未保存；保存不会自动连接或重启。';document.getElementById('connection-add-kind').value='wss';document.getElementById('connection-add-kind-note').textContent='仅支持 ws://127.0.0.1:明确端口和根路径，必须填写 Token；重启后在接入程序配置相同地址与 Token。';document.getElementById('connection-add-without-token').disabled=true;
    for (const id of ['connection-select','token-action','backup-scope']) {
      const select=document.getElementById(id); if (!select.children.length) select.add(new Option('启动配置 · 本机连接','fixture'));
    }
    document.getElementById('connection-draft-fields').hidden=false;document.getElementById('connection-draft-field-state').textContent='请求处理中，保留当前字段摘要，请等待结果；不会自动保存。';document.getElementById('connection-draft-field-list').innerHTML=['反向监听 9007199254740991','主人账号 · 启动配置','尚未保存的新连接'].map((title,i)=>'<article><h3>'+title+'</h3><p class="field-help">未保存字段：'+(i===0?2:i===1?1:3)+' 项（Token 操作、输入与确认合并计数）。</p><ul>'+(i===0?['WebSocket 地址','Token']:i===1?['主人 QQ 账号']:['连接方向','WebSocket 地址','Token']).map(label=>'<li>'+label+'</li>').join('')+'</ul></article>').join('');
    document.getElementById('plugin-scope').value='group';document.getElementById('plugin-group').value='1234567890123456789';document.getElementById('plugin-scope-state').textContent='当前范围：群 1234567890123456789。显式全局设置优先生效。';for(const id of ['plugin-scope','plugin-group','apply-plugin-scope','reload-plugins','discard-plugin-drafts','plugin-only-drafts','plugin-page-size'])document.getElementById(id).disabled=true;
    document.getElementById('plugin-response-panel').hidden=false;document.getElementById('plugin-response-title').textContent='响应状态 · 群 1234567890123456789';document.getElementById('plugin-response-state').textContent='全局响应：开启；此群单独响应：关闭；此群响应检查：通过。';document.getElementById('plugin-response-detail').textContent='响应来源：全局。此群单独关闭不能覆盖已开启的全局响应。';
    document.getElementById('plugin-batch-status').textContent='群 1234567890123456789：已确认成功 100 / 108 项，正在保存 long-plugin-identifier-for-save-all-progress-layout-testing-0123456789。';document.getElementById('save-all-plugin-drafts').disabled=true;document.getElementById('stop-plugin-batch').disabled=false;
    document.getElementById('plugin-page-status').textContent='第 6 / 6 页；显示 101～108 / 108 个匹配插件。';
    document.getElementById('plugin-catalog-source').textContent='清单范围：群 1234567890123456789；最近成功读取：2026/10/11 15:00:00。最近成功保存开关：2026/10/11 15:01:00（只更新已保存项）。最近读取失败，清单仍使用当前已加载范围的旧结果。不是服务端实时快照。';
    document.getElementById('plugin-catalog-status').textContent='已发起清单下载，共 108 个匹配插件（跨全部已加载页）。最近读取失败，使用当前已加载范围的旧结果。未保存标记不表示已生效；清单不能导入或恢复。';
    document.getElementById('plugin-list').innerHTML='<article class="plugin-card"><h2>long-plugin-identifier-for-layout-validation</h2><p>测试插件的长说明文字，用于检查窄窗口内的折行与开关控件。</p><footer><span class="plugin-default">当前生效：启用</span><button class="text-button">查看用法 →</button></footer><div class="plugin-controls"><select disabled><option>继承全局或默认</option></select><button class="secondary" disabled>保存开关</button></div><div class="plugin-single-draft"><p class="field-help plugin-draft-note">未保存</p><button class="secondary plugin-discard-one" disabled>放弃此项更改</button></div></article>';
    document.getElementById('ai-global-fields').disabled=false;
    document.getElementById('ai-config-checks').hidden=false;document.getElementById('ai-config-check-list').innerHTML=['聊天 / GPT','Agent','识图（可选）'].map(title=>'<article><h3>'+title+'</h3><p class="ai-check-result">基础字段待补充或检查；尚未验证实际调用</p><p class="field-help">接口类型：支持；模型：未填写（可选，按接口要求核对）；地址：需检查格式或隐藏认证参数；密钥：未配置（OLLaMA 可无密钥）。</p></article>').join('');
    document.getElementById('ai-draft-fields').hidden=false;document.getElementById('ai-draft-field-state').textContent='请求处理中，保留当前字段摘要，请等待结果；不会自动保存。';document.getElementById('ai-draft-field-list').innerHTML=['全局模型与人物设定 · 全局','聊天会话参数 · 私聊 QQ 1234567890123456789','GPT 温度 · 私聊 QQ 1234567890123456789'].map((title,i)=>'<article><h3>'+title+'</h3><p class="field-help">未保存字段：'+(i===0?20:i===1?5:1)+' 项（密钥操作、输入与确认合并计数）。</p><ul>'+(i===0?['聊天接口类型','聊天模型名','聊天接口地址','最大输出长度（0 为默认 4096）','TopP（0 为默认 0.9）','系统提示词','不发送系统提示词','推理努力度（DeepSeek 官方接口：none 关闭思考，low/high/max 开启；留空用模型默认）','回复分隔符（可留空）','识图接口类型','识图模型名（可留空）','识图接口地址（可留空）','Agent 接口类型','Agent 模型名（可留空）','Agent 接口地址（可留空）','Agent 人物设定（修改后保留对话记录）','Agent 性别','聊天密钥','识图密钥','Agent 密钥']:i===1?['温度百分比','非 @ 消息触发概率','Agent 模式','QQ AI 语音输出','响应 @ 消息']:['温度百分比']).map(label=>'<li>'+label+'</li>').join('')+'</ul></article>').join('');
    document.getElementById('ai-draft-status').textContent='未保存更改：全局模型与人物设定、聊天群参数、GPT / 群总结温度。请分别保存；切换后台页面会保留草稿。';
    document.getElementById('plugin-draft-status').textContent='未保存更改：108 个插件（群 1234567890123456789）。可单项保存或逐项保存全部草稿；搜索、筛选和切换页面会保留选择。';
    document.getElementById('ai-global-inputs').innerHTML='<label>模型请求地址</label><input value="http://127.0.0.1:9999/synthetic-model"><label>系统提示词</label><textarea rows="4">仅用于页面布局验证的合成内容。</textarea><label class="check-label"><input type="checkbox">确认清除密钥（未勾选）</label>';
    for (const [id,label,options] of [
      ...['Type','ImageType','AgentType'].map(name=>['ai-field-'+name,'接口类型',['OpenAI','OLLaMA','GenAI']]),
      ...['Key','ImageKey','AgentKey'].map(name=>['ai-action-'+name,'密钥操作',['保留当前密钥','替换为新密钥','清除密钥']])
    ]) {
      const caption=document.createElement('label'),select=document.createElement('select');caption.htmlFor=id;caption.textContent=label;select.id=id;
      for(const text of options){const option=document.createElement('option');option.textContent=text;select.append(option);}
      document.getElementById('ai-global-inputs').append(caption,select);
    }
    document.getElementById('ai-group-forms').innerHTML=['聊天会话参数 · 私聊 QQ 1234567890123456789','GPT / 群总结温度 · 群 1234567890123456789'].map(title=>'<form class="panel ai-form"><fieldset><h2>'+title+'</h2><label>温度百分比（0为默认70）</label><input type="number" value="70"><button type="submit" class="primary">备份并保存会话参数</button><button type="button" class="secondary ai-discard">放弃此表单更改</button></fieldset></form>').join('');
    document.getElementById('backup-list').innerHTML='<article class="panel backup-card"><div><h2>控制数据库</h2><p class="muted">测试备份 · 格式 v1 · 含校验摘要</p><small>20261003T000000.000000000Z-0123456789abcdef-control</small></div><button class="secondary">预览恢复</button></article>';
    document.getElementById('backup-size-summary').hidden=false;
    document.getElementById('backup-size-loaded').textContent='已加载：200 份；可核对内容大小：16.00 PiB（18014398509481982 字节）；大小无法核对：10 份。';
    document.getElementById('backup-size-matched').textContent='当前匹配（跨全部已加载页）：120 份；可核对内容大小：120.00 MiB（125829120 字节）；大小无法核对：5 份。';
    document.getElementById('backup-size-status').textContent='最近成功读取：2026/10/11 12:30:00。上次读取结果，刷新失败。仅加载最新200份，更早未加载记录不在摘要中。';
    document.getElementById('backup-size-scopes').innerHTML=['启动配置','控制数据库','后台启用与地址','异常/未知范围'].map(label=>'<article><h3>'+label+'</h3><p>已加载：200 份；可核对内容大小：16.00 PiB（18014398509481982 字节）；大小无法核对：10 份。</p><p>当前匹配：120 份；可核对内容大小：120.00 MiB（125829120 字节）；大小无法核对：5 份。</p></article>').join('');
    document.getElementById('backup-filter-count').textContent='200 / 200 份已加载备份；含摘要 180，旧版无摘要 10，不可恢复 10。仅检索本次加载的最新 200 份，更早未加载记录不在结果中。';
    document.getElementById('download-backup-catalog').disabled=false;document.getElementById('backup-export-status').textContent='已发起清单下载，共200份匹配记录（跨全部已加载页）。使用上次读取结果，最近刷新失败。更早未加载备份未列入。清单仅供核对，不能用于恢复。';
    document.getElementById('backup-page-status').textContent='第 10 / 10 页；显示 181～200 / 200 份匹配记录。';
    document.getElementById('restore-preview').hidden=false;
    document.getElementById('restore-backup-id').textContent='备份编号：20261009T000000.000000000Z-0123456789abcdef-control';
    document.getElementById('restore-expiry').textContent='预览有效至23:59:59；剩余9分59秒（本机估算）。';document.getElementById('restore-expiry-state').textContent='到期需重新预览；提交时仍会校验当前状态，不会自动恢复。';document.getElementById('cleanup-expiry').textContent='预览有效至23:59:59；剩余0分00秒。';document.getElementById('cleanup-expiry-state').textContent='预览已过期，请重新预览并核对后再次确认。';document.getElementById('cleanup-expiry-state').dataset.expired='true';
    document.getElementById('restore-detail').textContent='用于布局测试的模拟预览，不执行恢复。';
    document.getElementById('restore-impact-status').textContent='共4个类别，2个有变化；按预览时当前保存状态比较。';
    document.getElementById('restore-impact-list').innerHTML=['响应规则与AI扩展','全局用户封禁','插件开关继承标记','插件与其他库内数据'].map((label,i)=>'<article class="restore-impact-row"><strong>'+label+'</strong><span class="restore-impact-state'+(i<2?' impact-changed':'')+'">'+(i<2?'有变化':'一致')+'</span><p class="field-help">当前 100000条记录 → 备份 100000条记录'+(i<2?'（数量相同，内容不同）':'')+'</p></article>').join('');
    document.getElementById('cleanup-preview').hidden=false;
    document.getElementById('cleanup-detail').textContent='本次将删除 1 份旧备份，约 512 KiB；跳过 2 份无法安全确认的备份。永久删除不能撤销，当前配置与对话记录不受影响。';
    document.getElementById('cleanup-list').innerHTML='<li>控制数据库 · 2026/9/1 00:00:00 · 20260901T000000.000000000Z-0123456789abcdef-control</li>';
    document.getElementById('clear-log-display').disabled=false;document.getElementById('log-clear-status').textContent='仅清空本页显示，匹配统计和导出仍涵盖服务保留记录；重新加载后可读取近期日志。';
    document.getElementById('log-from').value='2026-10-09T16:00:00';document.getElementById('log-until').value='2026-10-09T17:00:00';
    document.getElementById('log-records').innerHTML='<article class="log-entry"><div class="log-meta"><time>2026/10/3 00:00:00</time><span class="log-level log-warning-level">WARNING</span></div><pre>本次启动记录；synthetic-long-line-for-wrap-testing-012345678901234567890123456789。</pre><div class="log-copy-actions"><button type="button" class="secondary copy-log-entry" disabled>复制本条</button><span class="field-help" role="status">复制未完成，已选中日志正文；请按 Ctrl+C 手动复制。</span></div></article>';
    document.getElementById('log-summary-state').textContent='符合已应用筛选：1000 / 服务仍保留：1000 条（最近成功读取于 20:00:00）。';document.getElementById('log-summary-levels').innerHTML=['Panic：125','Fatal：125','错误：125','警告：125','信息：125','调试：125','跟踪：125','其他：125'].map(text=>'<span>'+text+'</span>').join('');
    document.getElementById('log-filter-status').textContent='已应用：警告与错误；关键词“'+ 'a'.repeat(256)+'”。输入修改尚未应用；查看与导出仍使用上述条件。';
  });
  for (const width of [320,375,640,1280]) {
    await page.setViewportSize({width,height:900});
    for (const mode of ['light','dark']) {
      await page.evaluate(()=>{document.getElementById('auth').hidden=false; document.getElementById('workspace').hidden=true;});
      await page.selectOption('#auth-theme',mode);
      for (const view of ['auth','overview','connections','plugins','ai','logs','backups','security']) {
        await page.evaluate(view=>{
          document.getElementById('auth').hidden=view!=='auth'; document.getElementById('workspace').hidden=view==='auth';
          for (const section of document.querySelectorAll('.content>section')) section.hidden=section.id!==view+'-view';
          for (const button of document.querySelectorAll('.nav-button')) button.classList.toggle('active',button.dataset.view===view);
          const button=document.querySelector('.nav-button.active'); if (button) document.getElementById('breadcrumb').textContent='工作空间 / '+button.querySelector('span').textContent;
        },view);
        const result=await page.evaluate(()=>({width:innerWidth,scrollWidth:document.documentElement.scrollWidth,theme:document.documentElement.dataset.theme,
          namedNavigation:[...document.querySelectorAll('.nav-button span:not(.nav-count)')].every(span=>getComputedStyle(span).display!=='none'),
          countBadgeHidden:getComputedStyle(document.getElementById('nav-count')).display==='none',
          themeControlVisible:[...document.querySelectorAll('[data-theme-select]')].some(el=>el.getClientRects().length>0),
          inputBackground:getComputedStyle(document.getElementById('connection-url')).backgroundColor,
          pageBackground:getComputedStyle(document.body).backgroundColor,
          aiSelects:[...document.querySelectorAll('#ai-global-inputs select')].filter(el=>el.getClientRects().length).map(el=>({id:el.id,height:el.getBoundingClientRect().height,fontSize:parseFloat(getComputedStyle(el).fontSize),width:el.getBoundingClientRect().width,containerWidth:el.parentElement.clientWidth,background:getComputedStyle(el).backgroundColor,surface:getComputedStyle(el.closest('.panel')).backgroundColor}))}));
        measurements.push({width,mode,view,...result});
        if (result.scrollWidth>width+1 || result.theme!==mode || !result.themeControlVisible || (view!=='auth' && !result.namedNavigation) || (width<=640 && !result.countBadgeHidden)) throw new Error(`Layout check failed: ${JSON.stringify(measurements.at(-1))}`);
        if(view==='ai' && (result.aiSelects.length!==6 || result.aiSelects.some(el=>el.height<44 || (width<=640 && el.fontSize<16) || Math.abs(el.width-el.containerWidth)>1 || el.background!==el.surface))) throw new Error(`AI dropdown layout check failed: ${JSON.stringify(measurements.at(-1))}`);
        if (width===375 || width===1280) await page.screenshot({path:path.join(out,`${width}-${mode}-${view}.png`),fullPage:true});
      }
    }
  }
  passed=true;
} finally {
  await browser.close();
  const result={completedAt:new Date().toISOString(),passed,offline:true,networkRequestsAllowed:false,measurements};
  await fs.writeFile(path.join(out,'result.json'),JSON.stringify(result,null,2));
  console.log(JSON.stringify({passed,checks:measurements.length,artifactDirectory:out}));
}
