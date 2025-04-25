<script setup lang="ts">
  import { toast } from "@/components/ui/sonner";
  import MdiGithub from "~icons/mdi/github";
  import MdiDiscord from "~icons/mdi/discord";
  import MdiFolder from "~icons/mdi/folder";
  import MdiAccount from "~icons/mdi/account";
  import MdiAccountPlus from "~icons/mdi/account-plus";
  import MdiLogin from "~icons/mdi/login";
  import MdiArrowRight from "~icons/mdi/arrow-right";
  import MdiLock from "~icons/mdi/lock";
  import MdiMastodon from "~icons/mdi/mastodon";
  import { Card, CardContent, CardFooter, CardHeader, CardTitle } from "@/components/ui/card";
  import { Button } from "@/components/ui/button";
  import LanguageSelector from "~/components/App/LanguageSelector.vue";
  import { Tooltip, TooltipContent, TooltipTrigger, TooltipProvider } from "@/components/ui/tooltip";

  useHead({
    title: "Homebox | Organize and Tag Your Stuff",
  });

  definePageMeta({
    layout: "empty",
    middleware: [
      () => {
        const ctx = useAuthContext();
        if (ctx.isAuthorized()) {
          return "/home";
        }
      },
    ],
  });

  const ctx = useAuthContext();

  const api = usePublicApi();

  const { data: status } = useAsyncData(async () => {
    const { data } = await api.status();

    if (data.demo) {
      username.value = "demo@example.com";
      password.value = "demo";
    }
    return data;
  });

  whenever(status, status => {
    if (status?.demo) {
      email.value = "demo@example.com";
      loginPassword.value = "demo";
    }
  });

  const route = useRoute();
  const router = useRouter();

  const username = ref("");
  const email = ref("");
  const password = ref("");
  const canRegister = ref(false);
  const remember = ref(false);

  const groupToken = computed<string>({
    get() {
      const params = route.query.token;

      if (typeof params === "string") {
        return params;
      }

      return "";
    },
    set(v) {
      router.push({
        query: {
          token: v,
        },
      });
    },
  });

  async function registerUser() {
    loading.value = true;

    const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;

    if (!emailRegex.test(email.value)) {
      toast.error("Invalid email address");
      loading.value = false;
      return;
    }

    const { error } = await api.register({
      name: username.value,
      email: email.value,
      password: password.value,
      token: groupToken.value,
    });

    if (error) {
      toast.error("Problem registering user");
      return;
    }

    toast.success("User registered");

    loading.value = false;
    registerForm.value = false;
  }

  onMounted(() => {
    if (groupToken.value !== "") {
      registerForm.value = true;
    }
  });

  const loading = ref(false);
  const loginPassword = ref("");
  const redirectTo = useState("authRedirect");

  async function login() {
    loading.value = true;
    const { error } = await ctx.login(api, email.value, loginPassword.value, remember.value);

    if (error) {
      toast.error("Invalid email or password");
      loading.value = false;
      return;
    }

    toast.success("Logged in successfully");

    navigateTo(redirectTo.value || "/home");
    redirectTo.value = null;
    loading.value = false;
  }

  const [registerForm, toggleLogin] = useToggle();
</script>

<template>
  <div class="flex min-h-screen flex-col">
    <!-- TODO: when daisyui is removed then add -z-10 -->
    <div class="absolute top-0 min-w-full fill-primary">
      <div class="flex min-h-[20vh] flex-col bg-primary" />
      <svg
        class="fill-primary drop-shadow-xl"
        xmlns="http://www.w3.org/2000/svg"
        viewBox="0 0 1440 320"
        preserveAspectRatio="none"
      >
        <path
          fill-opacity="1"
          d="M0,32L80,69.3C160,107,320,181,480,181.3C640,181,800,107,960,117.3C1120,128,1280,224,1360,272L1440,320L1440,0L1360,0C1280,0,1120,0,960,0C800,0,640,0,480,0C320,0,160,0,80,0L0,0Z"
        ></path>
      </svg>
    </div>
    <div>
      <header class="mx-auto p-4 sm:flex sm:items-end sm:p-6 lg:p-14">
        <div class="z-10">
          <h2 class="mt-1 flex text-4xl font-bold tracking-tight text-neutral-content sm:text-5xl lg:text-6xl">
            HomeB
            <AppLogo class="-mb-4 w-12" />
            x
          </h2>
          <p class="ml-1 text-lg text-base-content/50">{{ $t("index.tagline") }}</p>
        </div>
        <TooltipProvider :delay-duration="0">
          <div class="z-10 ml-auto mt-6 flex items-center gap-4 text-neutral-content sm:mt-0">
            <Tooltip>
              <TooltipTrigger as-child>
                <a href="https://github.com/sysadminsmedia/homebox" target="_blank">
                  <MdiGithub class="size-8" />
                </a>
              </TooltipTrigger>
              <TooltipContent>{{ $t("global.github") }}</TooltipContent>
            </Tooltip>

            <Tooltip>
              <TooltipTrigger as-child>
                <a href="https://noc.social/@sysadminszone" target="_blank">
                  <MdiMastodon class="size-8" />
                </a>
              </TooltipTrigger>
              <TooltipContent>{{ $t("global.follow_dev") }}</TooltipContent>
            </Tooltip>

            <Tooltip>
              <TooltipTrigger as-child>
                <a href="https://discord.gg/aY4DCkpNA9" target="_blank">
                  <MdiDiscord class="size-8" />
                </a>
              </TooltipTrigger>
              <TooltipContent>{{ $t("global.join_discord") }}</TooltipContent>
            </Tooltip>

            <Tooltip>
              <TooltipTrigger as-child>
                <a href="https://homebox.software/en/" target="_blank">
                  <MdiFolder class="size-8" />
                </a>
              </TooltipTrigger>
              <TooltipContent>{{ $t("global.read_docs") }}</TooltipContent>
            </Tooltip>

            <LanguageSelector class="z-10 text-primary" :expanded="false" />
          </div>
        </TooltipProvider>
      </header>
      <div class="grid min-h-[50vh] p-6 sm:place-items-center">
        <!-- TODO: when daisyui is removed then remove z-10 -->
        <div class="z-10">
          <Transition name="slide-fade">
            <form v-if="registerForm" @submit.prevent="registerUser">
              <Card class="md:w-[500px]">
                <CardHeader>
                  <CardTitle class="flex items-center gap-2">
                    <MdiAccount class="mr-1 size-7" />
                    {{ $t("index.register") }}
                  </CardTitle>
                </CardHeader>
                <CardContent class="flex flex-col gap-2">
                  <FormTextField v-model="email" :label="$t('index.set_email')" />
                  <FormTextField v-model="username" :label="$t('index.set_name')" />
                  <div v-if="!(groupToken == '')" class="pb-1 pt-4 text-center">
                    <p>{{ $t("index.joining_group") }}</p>
                    <button type="button" class="text-xs underline" @click="groupToken = ''">
                      {{ $t("index.dont_join_group") }}
                    </button>
                  </div>
                  <FormPassword v-model="password" :label="$t('index.set_password')" />
                  <PasswordScore v-model:valid="canRegister" :password="password" />
                </CardContent>
                <CardFooter>
                  <Button
                    class="w-full"
                    type="submit"
                    :class="loading ? 'loading' : ''"
                    :disabled="loading || !canRegister"
                  >
                    {{ $t("index.register") }}
                  </Button>
                </CardFooter>
              </Card>
            </form>
            <form v-else @submit.prevent="login">
              <Card class="md:w-[500px]">
                <CardHeader>
                  <CardTitle class="flex items-center gap-2">
                    <MdiAccount class="mr-1 size-7" />
                    {{ $t("index.login") }}
                  </CardTitle>
                </CardHeader>
                <CardContent class="flex flex-col gap-2">
                  <template v-if="status && status.demo">
                    <p class="text-center text-xs italic">This is a demo instance</p>
                    <p class="text-center text-xs">
                      <b>{{ $t("global.email") }}</b> demo@example.com
                    </p>
                    <p class="text-center text-xs">
                      <b>{{ $t("global.password") }}</b> demo
                    </p>
                  </template>
                  <FormTextField v-model="email" :label="$t('global.email')" />
                  <FormPassword v-model="loginPassword" :label="$t('global.password')" />
                  <div class="max-w-[140px]">
                    <FormCheckbox v-model="remember" :label="$t('index.remember_me')" />
                  </div>
                </CardContent>
                <CardFooter>
                  <Button class="w-full" type="submit" :class="loading ? 'loading' : ''" :disabled="loading">
                    {{ $t("index.login") }}
                  </Button>
                </CardFooter>
              </Card>
            </form>
          </Transition>
          <div class="mt-6 text-center">
            <Button v-if="status && status.allowRegistration" class="group" variant="link" @click="() => toggleLogin()">
              <div class="relative mx-2">
                <div
                  class="absolute inset-0 flex items-center justify-center transition-transform duration-300 group-hover:rotate-[360deg]"
                >
                  <MdiAccountPlus v-if="!registerForm" class="size-5 group-hover:hidden" />
                  <MdiLogin v-else class="size-5 group-hover:hidden" />
                  <MdiArrowRight class="hidden size-5 group-hover:block" />
                </div>
              </div>
              {{ registerForm ? $t("index.login") : $t("index.register") }}
            </Button>
            <p v-else class="inline-flex items-center gap-2 text-sm italic text-base-content">
              <MdiLock class="inline-block size-4" />
              {{ $t("index.disabled_registration") }}
            </p>
          </div>
        </div>
      </div>
    </div>
    <footer v-if="status" class="bottom-0 mt-auto w-full pb-4 text-center">
      <p class="text-center text-sm">
        {{ $t("global.version", { version: status.build.version }) }} ~
        {{ $t("global.build", { build: status.build.commit }) }}
      </p>
    </footer>
  </div>
</template>

<style lang="css" scoped>
  .slide-fade-enter-active {
    transition: all 0.2s ease-out;
  }

  .slide-fade-enter-from,
  .slide-fade-leave-to {
    position: absolute;
    transform: translateX(20px);
    opacity: 0;
  }

  progress[value]::-webkit-progress-value {
    transition: width 0.5s;
  }
</style>
