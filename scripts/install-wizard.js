#!/usr/bin/env node
// Copyright (c) 2026 Lark Technologies Pte. Ltd.
// SPDX-License-Identifier: MIT

const { execFileSync, execSync } = require("child_process");
const p = require("@clack/prompts");

const PKG = "@larksuite/cli";
const SKILLS_REPO = "https://github.com/larksuite/cli";
const isWindows = process.platform === "win32";

// ---------------------------------------------------------------------------
// i18n
// ---------------------------------------------------------------------------

const messages = {
  zh: {
    setup:          "正在设置飞书 CLI...",
    step1:          "全局安装 %s",
    step1Skip:      "已安装 (v%s)，跳过。",
    step1Done:      "全局安装完成。",
    step1Fail:      "全局安装失败。请手动运行: npm install -g %s",
    step2:          "安装 AI Skills",
    step2Skip:      "已安装，跳过。",
    step2Spinner:   "安装 Skills 中...",
    step2Done:      "Skills 安装完成。",
    step2Fail:      "Skills 安装失败。请手动运行: npx skills add %s -y -g",
    step3:          "配置应用",
    step3NotFound:  "安装后未找到 lark-cli，终止。",
    step3Found:     "发现已配置应用 (App ID: %s)，是否继续使用？",
    step3Skip:      "跳过应用配置。",
    step3Done:      "应用配置完成。",
    step3Fail:      "应用配置失败。请手动运行: lark-cli config init --new",
    step4:          "用户授权",
    step4NotFound:  "未找到 lark-cli，跳过授权。",
    step4Confirm:   "是否允许 AI 访问你的飞书数据（消息、文档、日历等）？",
    step4Skip:      "跳过授权。后续可运行 lark-cli auth login 授权。",
    step4Done:      "用户授权完成。",
    step4Fail:      "授权失败，后续可重试: lark-cli auth login",
    done:           "安装完成！",
    doneHint:       "现在可以对你的 AI 工具（Claude Code、Trae 等）说：\n\"飞书 CLI 能帮我做什么？结合我的情况推荐一下从哪里开始\"",
    cancelled:      "安装已取消。",
  },
  en: {
    setup:          "Setting up Feishu/Lark CLI...",
    step1:          "Install %s globally",
    step1Skip:      "Already installed (v%s). Skipped.",
    step1Done:      "Global installation complete.",
    step1Fail:      "Failed to install globally. Try running manually: npm install -g %s",
    step2:          "Install AI Skills",
    step2Skip:      "Already installed. Skipped.",
    step2Spinner:   "Installing skills...",
    step2Done:      "Skills installation complete.",
    step2Fail:      "Skills installation failed. Try running manually: npx skills add %s -y -g",
    step3:          "Configure your app",
    step3NotFound:  "lark-cli not found after installation. Aborting.",
    step3Found:     "Found existing app (App ID: %s). Use this app?",
    step3Skip:      "Skipped app configuration.",
    step3Done:      "App configuration complete.",
    step3Fail:      "App configuration failed. Try running manually: lark-cli config init --new",
    step4:          "User authorization",
    step4NotFound:  "lark-cli not found. Skipping authorization.",
    step4Confirm:   "Allow AI to access your Feishu/Lark data (messages, docs, calendar, etc.)?",
    step4Skip:      "Skipped. You can run lark-cli auth login later.",
    step4Done:      "User authorization complete.",
    step4Fail:      "Authorization failed. You can retry later: lark-cli auth login",
    done:           "You're all set!",
    doneHint:       'Now try asking your AI tool (Claude Code, Trae, etc.):\n"What can Feishu/Lark CLI help me with, and where should I start?"',
    cancelled:      "Installation cancelled.",
  },
};

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

function handleCancel(value, msg) {
  if (p.isCancel(value)) {
    p.cancel(msg.cancelled);
    process.exit(0);
  }
  return value;
}

function execCmd(cmd, args, opts) {
  if (isWindows) {
    return execFileSync("cmd.exe", ["/c", cmd, ...args], opts);
  }
  return execFileSync(cmd, args, opts);
}

function run(cmd, args, opts = {}) {
  execCmd(cmd, args, { stdio: "inherit", ...opts });
}

function runSilent(cmd, args, opts = {}) {
  return execCmd(cmd, args, {
    stdio: ["ignore", "pipe", "pipe"],
    ...opts,
  });
}

function fmt(template, ...values) {
  let i = 0;
  return template.replace(/%s/g, () => values[i++] ?? "");
}

/** Resolve the path of globally installed lark-cli. */
function whichLarkCli() {
  try {
    const cmd = isWindows ? "where" : "which";
    return execSync(`${cmd} lark-cli`, { stdio: ["ignore", "pipe", "pipe"] })
      .toString()
      .split("\n")[0]
      .trim();
  } catch (_) {
    return null;
  }
}

/** Get the version of a lark-cli binary, or null. */
function getLarkCliVersion(binPath) {
  try {
    const out = runSilent(binPath, ["--version"], { timeout: 10000 });
    const match = out.toString().match(/(\d+\.\d+\.\d+)/);
    return match ? match[1] : null;
  } catch (_) {
    return null;
  }
}

/** Check whether lark-cli config already exists. Returns app ID or null. */
function getExistingAppId(binPath) {
  try {
    const out = runSilent(binPath, ["config", "show"], { timeout: 10000 });
    const json = JSON.parse(out.toString());
    return json.appId || null;
  } catch (_) {
    return null;
  }
}

/** Parse --lang from process.argv, returns "zh", "en", or null. */
function parseLangArg() {
  const args = process.argv.slice(2);
  for (let i = 0; i < args.length; i++) {
    if (args[i] === "--lang" && args[i + 1]) {
      const val = args[i + 1].toLowerCase();
      if (val === "zh" || val === "en") return val;
    }
    if (args[i].startsWith("--lang=")) {
      const val = args[i].split("=")[1].toLowerCase();
      if (val === "zh" || val === "en") return val;
    }
  }
  return null;
}

function skillsAlreadyInstalled() {
  try {
    const out = runSilent("npx", ["-y", "skills", "ls", "-g", "--json"], {
      timeout: 30000,
    });
    const list = JSON.parse(out.toString());
    return Array.isArray(list) && list.some((s) => s.name && s.name.startsWith("lark-"));
  } catch (_) {
    return false;
  }
}

// ---------------------------------------------------------------------------
// Steps
// ---------------------------------------------------------------------------

async function stepSelectLang() {
  const fromArg = parseLangArg();
  if (fromArg) return fromArg;

  const lang = await p.select({
    message: "请选择语言 / Select language",
    options: [
      { value: "zh", label: "中文" },
      { value: "en", label: "English" },
    ],
  });
  return handleCancel(lang, messages.zh);
}

async function stepInstallGlobally(msg) {
  const existing = whichLarkCli();
  if (existing) {
    const ver = getLarkCliVersion(existing);
    p.log.info(fmt(msg.step1Skip, ver || "unknown"));
    return;
  }

  const s = p.spinner();
  s.start(fmt(msg.step1, PKG));
  try {
    runSilent("npm", ["install", "-g", PKG], { timeout: 120000 });
    s.stop(msg.step1Done);
  } catch (_) {
    s.stop(fmt(msg.step1Fail, PKG));
    process.exit(1);
  }
}

async function stepInstallSkills(msg) {
  if (skillsAlreadyInstalled()) {
    p.log.info(msg.step2Skip);
    return;
  }

  const s = p.spinner();
  s.start(msg.step2Spinner);
  try {
    runSilent("npx", ["-y", "skills", "add", SKILLS_REPO, "-y", "-g"], {
      timeout: 120000,
    });
    s.stop(msg.step2Done);
  } catch (_) {
    s.stop(fmt(msg.step2Fail, SKILLS_REPO));
    process.exit(1);
  }
}

async function stepConfigInit(msg, lang) {
  const larkCli = whichLarkCli();
  if (!larkCli) {
    p.log.error(msg.step3NotFound);
    process.exit(1);
  }

  const appId = getExistingAppId(larkCli);
  if (appId) {
    const reuse = await p.confirm({
      message: fmt(msg.step3Found, appId),
    });
    if (handleCancel(reuse, msg) && reuse) {
      p.log.info(msg.step3Skip);
      return;
    }
  }

  p.log.step(msg.step3);
  try {
    run(larkCli, ["config", "init", "--new", "--lang", lang]);
    p.log.success(msg.step3Done);
  } catch (_) {
    p.log.error(msg.step3Fail);
    process.exit(1);
  }
}

async function stepAuthLogin(msg) {
  const larkCli = whichLarkCli();
  if (!larkCli) {
    p.log.warn(msg.step4NotFound);
    return;
  }

  const yes = await p.confirm({
    message: msg.step4Confirm,
  });
  if (p.isCancel(yes)) {
    p.cancel(msg.cancelled);
    process.exit(0);
  }
  if (!yes) {
    p.log.info(msg.step4Skip);
    return;
  }

  p.log.step(msg.step4);
  try {
    run(larkCli, ["auth", "login"]);
    p.log.success(msg.step4Done);
  } catch (_) {
    p.log.warn(msg.step4Fail);
  }
}

// ---------------------------------------------------------------------------
// Main
// ---------------------------------------------------------------------------

async function main() {
  const lang = await stepSelectLang();
  const msg = messages[lang];

  p.intro(msg.setup);

  await stepInstallGlobally(msg);
  await stepInstallSkills(msg);
  await stepConfigInit(msg, lang);
  await stepAuthLogin(msg);

  p.outro(msg.done);
  console.log(msg.doneHint);
}

main().catch((err) => {
  p.cancel("Unexpected error: " + (err.message || err));
  process.exit(1);
});
