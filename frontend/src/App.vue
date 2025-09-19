<template>
  <main :style="mainStyle">
    <MainWindow />

    <LoginWindow />
    <RegisterWindow />
    <ProfileWindow />
  </main>
</template>

<script setup>
import MainWindow from './components/MainWindow.vue';
import LoginWindow from './components/LoginWindow.vue';
import RegisterWindow from './components/RegisterWindow.vue';
import ProfileWindow from './components/ProfileWindow.vue';
import background from './assets/windows7.jpg';

const mainStyle = `background-image: url(${background})`;
</script>

<style scoped>
main {
  min-height: 100vh;
  display: grid;
  place-items: center;
  text-align: center;
  background-position: center;
  background-size: cover;
  background-repeat: no-repeat;
}
</style>
