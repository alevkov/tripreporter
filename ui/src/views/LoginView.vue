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


<script setup>
import LoginForm from "@/components/LoginForm.vue";
import AlreadyLoggedIn from "@/components/AlreadyLoggedIn.vue";
import SubjectiveReportLink from "@/components/SubjectiveReportLink.vue";
import { ref, inject } from 'vue';  // Import from 'vue' instead of '@vue/runtime-core'
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
    success.value = response.status === 200;
    submitting.value = false;
    setMessage(response.data.msg, messageSuccess, success.value, undefined, "/login");
  } catch (error) {  // 'error' is now defined as 'err'
    success.value = error.response && error.response.status === 200;
    submitting.value = false;
    setMessage(error.response.data.msg, messageSuccess, success.value, undefined, "/login");
    handleMessageError(error)
  }

  if (success.value) {
    store.updateSession(axios);
  }
};
</script>

<script>
export default {
  components: {
    LoginForm,
    AlreadyLoggedIn,
    SubjectiveReportLink
  }
};
</script>

<style>
@import url(@/assets/css/forms.css);
</style>

<style scoped>
@import url(@/assets/css/message_util.css);
</style>
