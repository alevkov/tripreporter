<template>
    <div class="DefaultView__form">
      <FormKit type="form" @submit="submitForm" #default="{ state: { errors } }" :actions="false">
        <FormKit
            type="text"
            name="username"
            id="username"
            label="Username"
            validation="required"
            placeholder="lyv76"
            :value="lastUsername"
        />
  
        <FormKit
            type="password"
            name="password"
            id="password"
            label="Password"
            validation="required"
            placeholder="----------"
        />
  
        <FormKit type="submit" label="Login" data-next="true" :disabled="errors && submitting"/>
      </FormKit>
    </div>
  </template>
  
  <script setup>

  import { defineProps, ref } from 'vue'
  
  const props = defineProps({
  onSubmit: Function,
  lastUsername: String
})

  
  const submitting = ref(false) 
  
  const submitForm = async (fields) => {
    submitting.value = true
    await props.onSubmit(fields) 
    submitting.value = false
  }
  
  </script>
  
  
  <style>
  @import url(@/assets/css/forms.css);
  </style>
  
  <style scoped>
  @import url(@/assets/css/message_util.css);
  </style>