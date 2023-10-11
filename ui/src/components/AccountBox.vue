<!--
SPDX-FileCopyrightText: 2023 froggie <legal@frogg.ie>

SPDX-License-Identifier: OSL-3.0
-->

<template>
  <h1 v-if="!getStore().showDeleteForm" class="--tr-header-h1">Manage Account</h1>
  <h1 v-else class="--tr-header-h1">Delete Account</h1>

  <div class="LayoutAccount__main" v-if="isLoaded()">
    <div v-show="!getStore().showDeleteForm">
      <div v-for="(account, index) in [getStore().accountJson]" :key="index">
        <div class="LayoutAccount__account">
          <div class="LayoutAccount__info">
            <div class="LayoutAccount__info_entry">
              <!-- TODO: Add edit buttons -->
              <header-row-box
                  class="LayoutAccount__info_entry_box"
                  style="margin-left: 0;"
                  header="Account"
                  icon="user"
                  :columns="['Profile', 'Username', 'Email']"
                  :rows="{
                    'Profile': getStore().accountJson.default_name,
                    'Username': getStore().accountJson.username,
                    'Email': getStore().accountJson.email,
                  }"
                  :links="{
                    'Profile': `/profile/${getStore().accountJson.id}`
                  }"
              />
              <div class="LayoutAccount__personal_details">
                <h2>Update Personal Details</h2>
                <FormKit type="form" name="subject_info" v-model="getCreateStore().reportSubject">
              <FormKit
                  type="number"
                  min="0"
                  id="subject_age"
                  name="subject_age"
                  label="Age"
                  help="(optional)"
                  v-model="getCreateStore().subjectInfo.age"
              />

              <FormKit
                  type="taglist"
                  id="subject_gender"
                  name="subject_gender"
                  label="Gender"
                  :options="getOptionsGender()"
                  :allow-new-values="true"
                  max="1"
                  help="(optional) Custom values are supported."
                  v-model="getCreateStore().subjectInfo.gender"
              />

              <FormKit
                  type="toggle"
                  name="use_imperial"
                  label="Use imperial for units?"
                  v-model="getCreateStore().useImperial"
              />

              <div v-show="!getCreateStore().useImperial">
                <FormKit
                    type="number"
                    min="0"
                    step="0.5"
                    id="subject_height_cm"
                    name="subject_height_cm"
                    label="Height (cm)"
                    help="(optional)"
                    v-model="getCreateStore().subjectInfo.heightCm"
                />
                <FormKit
                    type="number"
                    min="0"
                    step="0.5"
                    id="subject_weight_kg"
                    name="subject_weight_kg"
                    label="Weight (kg)"
                    help="(optional)"
                    v-model="getCreateStore().subjectInfo.weightKg"
                />
              </div>
              <div v-show="getCreateStore().useImperial">
                <FormKit
                    type="number"
                    min="0"
                    id="subject_height_ft"
                    name="subject_height_ft"
                    label="Height (ft)"
                    help="(optional)"
                    v-model="getCreateStore().subjectInfo.heightFt"
                />
                <FormKit
                    type="number"
                    min="0"
                    step="0.5"
                    id="subject_height_in"
                    name="subject_height_in"
                    label="Height (in)"
                    help="(optional)"
                    v-model="getCreateStore().subjectInfo.heightIn"
                />
                <FormKit
                    type="number"
                    min="0"
                    step="0.5"
                    id="subject_weight_lbs"
                    name="subject_weight_lbs"
                    label="Weight (lbs)"
                    help="(optional)"
                    v-model="getCreateStore().subjectInfo.weightLbs"
                />
              </div>
            </FormKit>
              </div>

              <div class="LayoutAccount_buttons">
                
                <FormKit
                    type="button"
                    class="LayoutAccount__buttons_button"
                    label="Logout"
                    @click="logOut"
                    :disabled="getSession().activeSession === false"
                />
                <FormKit
                    type="button"
                    style="background-color: var(--tr-error)"
                    class="LayoutAccount__buttons_button"
                    label="Delete Account"
                    @click="showDeleteForm"
                    :disabled="getSession().activeSession === false"
                />
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
    <div v-show="getStore().showDeleteForm">

      <div class="DefaultView__form">
        <FormKit type="form" @submit="submitForm" #default="{ state: { valid } }" :actions="false">
          <FormKit
              type="email"
              name="email"
              id="email"
              label="Email address"
              help="The email associated with your account."
              validation="required|email"
              placeholder="lyv@effectindex.com"
          />

          <FormKit
              type="password"
              name="password"
              id="password"
              label="Password"
              help="The password associated with your account."
              validation="required|length:8,32"
              placeholder="----------"
          />

          <div class="LayoutAccount_buttons">
            <FormKit
                type="button"
                class="LayoutAccount__buttons_button"
                label="Back"
                @click="hideDeleteForm"
                :disabled="getSession().activeSession === false"
            />
            <FormKit
                type="submit"
                style="background-color: var(--tr-error)"
                class="LayoutAccount__buttons_button"
                label="Delete Account"
                data-next="true"
                :disabled="!valid || getSession().activeSession === false"
            />
          </div>
        </FormKit>

          
      </div>
    </div>
  </div>
</template>

<script>
import HeaderRowBox from "@/components/HeaderRowBox.vue";
import { useAccountStore } from "@/assets/lib/accountstore";
import { useSessionStore } from "@/assets/lib/sessionstore";
import { useCreateStore } from "@/assets/lib/createstore";
const optionsGender = ["Male", "Female", "Nonbinary"];

const store = useAccountStore();
const sessionStore = useSessionStore();
const createStore = useCreateStore()

export default {
  name: "AccountBox",
  components: { HeaderRowBox },
  methods: {
    getStore() {
      return store
    },
    getSession() {
      return sessionStore
    },
    isLoaded() {
      return store.isLoaded()
    },
    getOptionsGender() {
      return optionsGender
    },
    getCreateStore() {
      return createStore
    }
  }
}
</script>

<script async setup>
import { inject, ref } from "vue";
import { handleMessageError, setMessage } from "@/assets/lib/message_util";
import log from "@/assets/lib/logger";

const router = inject('router')
const axios = inject('axios')

const messageSuccess = "Account successfully deleted!<br>You will be redirected to the home page in 3 seconds.";
const messageLoggedOut = "Logged out successfully!<br>You will be redirected to login in 3 seconds.";
let success = ref(false);
let loggedOut = ref(false);
let submitting = ref(false);
let ranSetup = false;

const submitForm = async (fields) => {
  log("submitForm", fields)

  submitting.value = true;

  axios.delete('/account', { data: fields }).then(function (response) {
    success.value = response.status === 200;
    submitting.value = false;
    if (success.value === true) {
      store.showDeleteForm = false;
      sessionStore.invalidateSession();
    }
    setMessage(response.data.msg, messageSuccess, success.value, router, '/', 3000);
  }).catch(function (error) {
    success.value = error.response.status === 200;
    submitting.value = false;
    if (success.value === true) {
      store.showDeleteForm = false;
      sessionStore.invalidateSession();
    }
    setMessage(error.response.data.msg, messageSuccess, success.value, router, '/', 3000);
    handleMessageError(error);
  })
}

const showDeleteForm = (e) => {
  store.showDeleteForm = true
}

const hideDeleteForm = (e) => {
  store.showDeleteForm = false
}

const logOut = (e) => {
  submitting.value = true

  axios.delete('/session').then(function (response) {
    loggedOut.value = response.status === 200;
    submitting.value = false;
    if (loggedOut.value === true) {
      store.showDeleteForm = false;
      sessionStore.invalidateSession();
    }
    setMessage(response.data.msg, messageLoggedOut, loggedOut.value, router, '/login', 3000);
  }).catch(function (error) {
    loggedOut.value = error.response.status === 200;
    submitting.value = false;
    if (loggedOut.value === true) {
      store.showDeleteForm = false;
      sessionStore.invalidateSession();
    }
    setMessage(error.response.data.msg, messageLoggedOut, loggedOut.value, router, '/login', 3000);
    handleMessageError(error);
  })
}

if (ranSetup !== true) {
  ranSetup = true

  await axios.get('/account').then(function (response) {
    store.updateData(response.status, response.data)
    setMessage(response.data.msg, "", store.apiSuccess);
  }).catch(function (error) {
    log("error", error)
    store.updateData(error.response.status, error.response.data)
    setMessage(error.response.data.msg, "", store.apiSuccess);
    handleMessageError(error);
  })
}
</script>

<style scoped>
.LayoutAccount__main {
  text-align: left;
}

.LayoutAccount__main h1 {
  text-align: center;
}

.LayoutAccount__account,
.LayoutAccount__personal_details {
  max-width: 25em;
  margin: auto;
}

.LayoutAccount__info_entry_box,
.LayoutAccount__personal_details {
  display: flex;
  flex-direction: column;
  font-size: 1em;
}

.LayoutAccount__personal_details h2 {
  font-size: 1.2em;
  align-self: center;
}

.LayoutAccount_buttons {
  display: flex;
  flex-wrap: wrap;
  flex-direction: row;
  justify-content: space-between;
}

.LayoutAccount__buttons_button {
  flex-grow: 1;
}

/* Additional style for Delete Account button */
.LayoutAccount__buttons_button--delete {
  background-color: var(--tr-error);
}
</style>


