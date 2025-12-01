<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, reactive, ref } from 'vue';
import { EventsOff, EventsOn } from '../wailsjs/runtime/runtime';
import { GetConfig, GetStatus, LoginNow, LogoutNow, SaveConfig } from '../wailsjs/go/main/App';

type Status = {
  online?: boolean;
  message?: string;
  last_check?: string;
  LastCheck?: string;
};

type Account = {
  StudentID?: string;
  Password?: string;
  Carrier?: string;
};

type Config = {
  WifiSSID?: string;
  AutoLoginInterval?: number;
  LoginMode?: string;
  Account?: Account;
  AutoStart?: boolean;
  OpenSettingsOnRun?: boolean;
};

const cfg = ref<Config | null>(null);
const status = ref<Status>({ online: false, message: '初始化中' });
const form = reactive({
  studentId: '',
  password: '',
  carrier: 'telecom',
  loginMode: 'operator_id',
  interval: 10,
});

const saving = ref(false);
const loading = ref(false);
let timer: number | undefined;

const lastCheckText = computed(() => {
  const raw = status.value.last_check || status.value.LastCheck;
  if (!raw) return '—';
  const ts = typeof raw === 'string' ? raw : String(raw);
  const parsed = new Date(ts);
  if (Number.isNaN(parsed.getTime())) return ts;
  return parsed.toLocaleString();
});

const statusText = computed(() => status.value.message || '未检测');

function applyConfigToForm(c: Config) {
  form.studentId = c.Account?.StudentID || '';
  form.password = c.Account?.Password || '';
  form.carrier = c.Account?.Carrier || 'telecom';
  form.loginMode = c.LoginMode || 'operator_id';
  form.interval = c.AutoLoginInterval && c.AutoLoginInterval > 0 ? c.AutoLoginInterval : 10;
}

async function loadConfig() {
  loading.value = true;
  try {
    const data = await GetConfig();
    const loaded: Config = (data || {}) as Config;
    cfg.value = loaded;
    applyConfigToForm(loaded);
  } finally {
    loading.value = false;
  }
}

async function saveConfig() {
  saving.value = true;
  try {
    const next: Config = cfg.value ? { ...cfg.value } : {};
    next.Account = next.Account || {};
    next.Account.StudentID = form.studentId;
    next.Account.Password = form.password;
    next.Account.Carrier = form.carrier;
    next.LoginMode = form.loginMode;
    next.AutoLoginInterval = form.interval;
    cfg.value = next;
    await SaveConfig(next as any);
  } finally {
    saving.value = false;
  }
}

async function refreshStatus() {
  try {
    status.value = await GetStatus();
  } catch (e) {
    console.error(e);
  }
}

async function loginNow() {
  try {
    await LoginNow();
  } catch (e) {
    console.error(e);
  }
}

async function logoutNow() {
  try {
    await LogoutNow();
  } catch (e) {
    console.error(e);
  }
}

onMounted(async () => {
  await loadConfig();
  await refreshStatus();
  timer = window.setInterval(refreshStatus, 1000);
  EventsOn('status:update', (st: Status) => {
    status.value = st;
  });
});

onBeforeUnmount(() => {
  if (timer) {
    window.clearInterval(timer);
  }
  EventsOff('status:update');
});
</script>

<template>
  <main class="page">
    <section class="panel">
      <header class="panel__header">
        <div>
          <p class="eyebrow">CUMT 校园网自动登录</p>
          <h1>登录配置</h1>
          <p class="muted">学号、后缀、登录模式与自动重试周期</p>
        </div>
        <div class="status-chip" :class="{ online: status?.online }">
          <span class="dot" />
          <span>{{ statusText }}</span>
        </div>
      </header>

      <div class="grid">
        <div class="card">
          <label class="field">
            <span>学号</span>
            <input v-model="form.studentId" autocomplete="off" placeholder="0823xxxx" />
          </label>

          <label class="field">
            <span>密码</span>
            <input v-model="form.password" type="password" autocomplete="off" placeholder="••••••" />
          </label>

          <label class="field">
            <span>运营商后缀</span>
            <select v-model="form.carrier">
              <option value="telecom">@telecom（电信）</option>
              <option value="cmcc">@cmcc（移动）</option>
              <option value="unicom">@unicom（联通）</option>
              <option value="none">无后缀</option>
            </select>
          </label>

          <label class="field">
            <span>登录模式</span>
            <select v-model="form.loginMode">
              <option value="operator_id">运营商账号（学号+后缀）</option>
              <option value="campus_only">校园网账号（纯学号）</option>
            </select>
          </label>

          <label class="field">
            <span>自动登录间隔（秒）</span>
            <input v-model.number="form.interval" type="number" min="5" />
          </label>
        </div>

        <div class="card">
          <div class="status-block">
            <p class="eyebrow">网络状态</p>
            <h2>{{ status?.online ? '已在线' : '离线' }}</h2>
            <p class="muted">{{ statusText }}</p>
            <p class="muted">最近检测：{{ lastCheckText }}</p>
          </div>

          <div class="actions">
            <button class="btn primary" type="button" @click="loginNow">立即登录</button>
            <button class="btn ghost" type="button" @click="logoutNow">注销</button>
            <button class="btn" type="button" :disabled="saving" @click="saveConfig">
              {{ saving ? '保存中...' : '保存配置' }}
            </button>
          </div>
        </div>
      </div>

      <footer class="panel__footer">
        <span>{{ loading ? '正在读取配置...' : '后台每秒检测一次网络状态' }}</span>
        <span>托盘：可右键打开设置、登录/注销、退出</span>
      </footer>
    </section>
  </main>
</template>
