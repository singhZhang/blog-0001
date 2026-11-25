<script setup lang="ts">
import { ref } from 'vue'
import { ArrowRight, CircleCheck } from '@lucide/vue'
import Icon from './ui/Icon.vue'

const email = ref('')
const message = ref('')
const submitted = ref(false)

function subscribe() {
  message.value = ''
  if (!/^\S+@\S+\.\S+$/.test(email.value)) {
    message.value = '请输入有效的邮箱地址。'
    return
  }
  submitted.value = true
}
</script>

<template>
  <section class="newsletter" aria-labelledby="newsletter-heading">
    <div>
      <p class="section-index">02 / NEWSLETTER</p>
      <h2 id="newsletter-heading">订阅更新，不错过新的技术记录</h2>
      <p>每月分享工程实践、系统设计与产品思考，一起构建更好的技术世界。</p>
    </div>
    <div v-if="submitted" class="success-message" role="status">
      <Icon :icon="CircleCheck" :size="24" />
      <span>订阅成功，下一封技术记录见。</span>
    </div>
    <form v-else novalidate @submit.prevent="subscribe">
      <label class="sr-only" for="newsletter-email">邮箱地址</label>
      <input id="newsletter-email" v-model.trim="email" type="email" autocomplete="email" placeholder="输入你的邮箱地址" :aria-describedby="message ? 'newsletter-error' : undefined" />
      <button type="submit">
        订阅更新
        <Icon :icon="ArrowRight" :size="19" />
      </button>
      <p v-if="message" id="newsletter-error" class="form-message" role="alert">{{ message }}</p>
    </form>
  </section>
</template>

<style scoped>
.newsletter {
  width: min(calc(100% - 48px), var(--content-wide));
  min-height: 124px;
  display: grid;
  grid-template-columns: 1fr minmax(390px, 520px);
  align-items: center;
  gap: 48px;
  margin: 42px auto 0;
  padding: 26px 34px;
  background: #dfeafe;
  color: #0b234f;
}

.section-index {
  margin: 0 0 7px;
  color: #4d648c;
  font-family: var(--font-mono);
  font-size: 10px;
  letter-spacing: 0.12em;
}

h2 {
  margin: 0 0 5px;
  font-size: clamp(21px, 2vw, 28px);
  letter-spacing: -0.045em;
}

.newsletter > div > p:last-child {
  margin: 0;
  color: #5f6f8d;
  font-size: 13px;
  line-height: 1.6;
}

form {
  position: relative;
  display: grid;
  grid-template-columns: 1fr auto;
}

input {
  min-width: 0;
  height: 50px;
  padding: 0 16px;
  border: 1px solid #2b6ef2;
  border-right: 0;
  border-radius: 0;
  background: #fff;
  color: var(--ink);
  font: inherit;
}

button {
  height: 50px;
  display: inline-flex;
  align-items: center;
  gap: 10px;
  padding: 0 24px;
  border: 1px solid var(--blue);
  background: var(--blue);
  color: #fff;
  font: inherit;
  font-weight: 700;
  cursor: pointer;
}

.form-message {
  position: absolute;
  top: calc(100% + 7px);
  left: 0;
  margin: 0;
  color: #972f2f;
  font-size: 12px;
}

.success-message {
  display: flex;
  align-items: center;
  justify-content: flex-end;
  gap: 10px;
  color: #174da3;
  font-weight: 700;
}

@media (max-width: 900px) {
  .newsletter {
    grid-template-columns: 1fr;
    gap: 22px;
  }
}

@media (max-width: 600px) {
  .newsletter {
    width: calc(100% - 24px);
    margin-top: 24px;
    padding: 24px 20px 30px;
  }
  form { grid-template-columns: 1fr; }
  input { border-right: 1px solid #2b6ef2; }
  button { justify-content: center; }
}
</style>
