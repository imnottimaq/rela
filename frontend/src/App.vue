<template>
          <!-- :buttons="[
        { label: 'Minimize', onClick: test },
        { label: 'Maximize', onClick: test },
        { label: 'Close', onClick: test }
      ]" -->
  <main :style="mainStyle">
    <WindowComponent
      title="Welcome to Rela"
      :menu="windowMenu"
      :initial-position="{ x: 160, y: 120 }"
      :initial-size="{ width: 640, height: 460 }"
      :visible="true"
    >
      <p>Rela is WIP task tracker with ability to self-host it.<br />Use Account to login or register.</p>
      <img class="demo" src="./assets/confused_travolta.gif" alt="Confused Travolta" />
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

const windowMenu = [
//   {
//     label: "Workspaces",
//     items: [
//       { type: "button", label: "New Workspace", onClick: test, divider: true },
//       { type: "button", label: "Workspace 1", onClick: test },
//       { type: "button", label: "Workspace 2", onClick: test },
//       { type: "button", label: "Workspace 3", onClick: test },
//     ],
//   },
  {
    label: "Account",
    items: [
      { type: "button", label: "Login", onClick: test },
      { type: "button", label: "Register", onClick: test, divider: true },
      { type: "button", label: "Logout", onClick: test },
    ],
  },
];
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
.demo {
  max-width: 100%;
  max-height: 70%;
  width: auto;
  height: auto;
  display: block;
  object-fit: contain;
  margin: 0 auto;
}

</style>
