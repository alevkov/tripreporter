<template>
  <div class="login">
    <div v-if="!store.activeSession" class="no-session">
      <h1 class="--tr-header-h1">Login to your
        <subjective-report-link/>
        account ✨
      </h1>
      <LoginForm :onSubmit="handleLogin" :lastUsername="lastUsername" />
    </div>
    <div v-else>
      <already-logged-in/>
    </div>
  </div>
</template>

<script>
import LayoutDefault from "@/layouts/LayoutDefault.vue";

export default {
  name: "LoginView",
  created() {
    console.log("this herer")
    this.$emit('update:layout', LayoutDefault);
  },
  components: {
    LoginForm,
    AlreadyLoggedIn,
    SubjectiveReportLink
  },
}
</script>

<script setup>
import {  ref, inject } from 'vue';
import LoginForm from "@/components/LoginForm.vue";
import AlreadyLoggedIn from "@/components/AlreadyLoggedIn.vue";
import SubjectiveReportLink from "@/components/SubjectiveReportLink.vue";
import { useSessionStore } from '@/assets/lib/sessionstore';
import { handleMessageError, setMessage } from '@/assets/lib/message_util';

const axios = inject('axios');
const store = useSessionStore();

const messageSuccess = "Successfully logged in!";
const success = ref(false);
const submitting = ref(false);

const handleLogin = async (fields) => {
  store.lastUsername = fields.username;
  submitting.value = true;

  try {
    const response = await axios.post('/account/login', fields);
    console.log(response)
    success.value = response.status === 200;
    submitting.value = false;
    setMessage(response.data.msg, messageSuccess, success.value, undefined, "/login");
  } catch (error) {
    success.value = error.response && error.response.status === 200;
    submitting.value = false;
    setMessage(error.response.data.msg, messageSuccess, success.value, undefined, "/login");
    handleMessageError(error);
  }

  if (success.value) {
    store.updateSession(axios);
  }
};
</script>

<style>
@import url(@/assets/css/forms.css);
</style>

<style scoped>
@import url(@/assets/css/message_util.css);
</style>
