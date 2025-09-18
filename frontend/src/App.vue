<template>
  <main :style="mainStyle">
    <WindowComponent 
    title="Welcome to Rela"
    :buttons="[
      { label: 'Minimize', onClick: test },
      { label: 'Maximize', onClick: test },
      { label: 'Close', onClick: test }
    ]"
    >
      <p>Lorem ipsum dolor sit amet, consectetur adipiscing elit. Sed eleifend velit lectus, in tempor tellus tincidunt sit amet. Sed eget viverra arcu. Nunc non velit tempus, consequat ligula ut, luctus quam. Phasellus consequat quam et ornare bibendum. Sed rutrum fermentum ipsum, a tincidunt libero blandit quis. Aliquam commodo nulla sapien, eget convallis nibh euismod at. Integer pulvinar tincidunt justo, vel porta nisi ullamcorper et. Vivamus dapibus tortor ut sapien maximus, in iaculis lectus viverra. Nullam faucibus lacinia arcu non vehicula. Nam nec accumsan erat. </p>
    </WindowComponent>
  </main>
</template>

<script setup>
import WindowComponent from './components/WindowComponent.vue';
import background from './assets/background.png';

const mainStyle = `background-image: url(${background})`;

const test = () => {
  alert("Test function");
};
</script>

<style scoped>
main {
  min-height: 100vh;
  display: grid;
  place-items: center;
  font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', sans-serif;
  text-align: center;
  background-position: center;
  background-size: cover;
  background-repeat: no-repeat;
}

</style>
