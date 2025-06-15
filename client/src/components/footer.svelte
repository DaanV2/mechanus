<script>
  import {
    Footer,
    FooterCopyright,
    FooterIcon,
    FooterLink,
    FooterLinkGroup
  } from 'flowbite-svelte';
  import { GithubSolid } from 'flowbite-svelte-icons';
</script>

<Footer
  footerType="socialmedia"
  class="bg-secondary-500 border-t-secondary-300 border-1 mt-auto flex flex-col py-2"
>
  <div class="grid gap-12 lg:grid-cols-6 lg:gap-24">
    <div>
      <h2 class="mb-6 text-sm font-semibold uppercase text-gray-400 dark:text-white">
        HELP & SUPPORT
      </h2>
      <FooterLinkGroup>
        <FooterLink liClass="mb-4" href="https://github.com/DaanV2/mechanus">GitHub</FooterLink>
      </FooterLinkGroup>
    </div>
    <div>
      <h2 class="mb-6 text-sm font-semibold uppercase text-gray-400 dark:text-white">Follow us</h2>
      <FooterLinkGroup>
        <FooterLink liClass="mb-4" href="https://github.com/DaanV2/mechanus">GitHub</FooterLink>
      </FooterLinkGroup>
    </div>
    <div>
      <h2 class="mb-6 text-sm font-semibold uppercase text-gray-400 dark:text-white">Legal</h2>
      <FooterLinkGroup>
        <FooterLink liClass="mb-4" href="https://github.com/DaanV2/mechanus/blob/main/License"
          >LICENSE</FooterLink
        >
      </FooterLinkGroup>
    </div>
  </div>
  <hr class="my-6 border-white sm:mx-auto lg:my-8 dark:border-white" />
  <div class="sm:flex sm:items-center sm:justify-between">
    <FooterCopyright href="https://github.com/DaanV2/mechanus" by="DaanV2" />
    <div class="mt-4 flex space-x-6 sm:mt-0 sm:justify-center">
      <FooterIcon
        href="https://github.com/DaanV2/mechanus"
        class="text-gray-400 hover:text-gray-900"><GithubSolid /></FooterIcon
      >
    </div>
  </div>
</Footer>
