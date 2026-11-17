<script>
  import {
    Footer,
    FooterCopyright,
    FooterIcon,
    FooterLink,
    FooterLinkGroup
  } from 'flowbite-svelte';
  import { GithubSolid } from 'flowbite-svelte-icons';

  const heading = 'mb-4 text-sm font-semibold text-gray-100 uppercase';
  const link = 'text-gray-300 hover:text-white hover:underline';
</script>

<!-- Same dark colours in both themes; `mt-auto` keeps it at the bottom of short pages -->
<Footer
  footerType="socialmedia"
  class="bg-secondary dark:bg-primary-950 border-line mt-auto border-t px-4 py-6 sm:px-6 lg:px-8"
>
  <div class="mx-auto w-full max-w-6xl">
    <div class="grid grid-cols-3 gap-4 sm:gap-8">
      <div>
        <h2 class={heading}>Help & support</h2>
        <FooterLinkGroup class="text-gray-300">
          <FooterLink liClass="mb-2" aClass={link} href="https://github.com/DaanV2/mechanus"
            >GitHub</FooterLink
          >
        </FooterLinkGroup>
      </div>
      <div>
        <h2 class={heading}>Follow us</h2>
        <FooterLinkGroup class="text-gray-300">
          <FooterLink liClass="mb-2" aClass={link} href="https://github.com/DaanV2/mechanus"
            >GitHub</FooterLink
          >
        </FooterLinkGroup>
      </div>
      <div>
        <h2 class={heading}>Legal</h2>
        <FooterLinkGroup class="text-gray-300">
          <FooterLink
            liClass="mb-2"
            aClass={link}
            href="https://github.com/DaanV2/mechanus/blob/main/LICENSE">LICENSE</FooterLink
          >
        </FooterLinkGroup>
      </div>
    </div>
    <hr class="my-6 border-gray-600" />
    <div class="flex items-center justify-between">
      <FooterCopyright
        class="text-gray-300"
        aClass="hover:text-white hover:underline"
        href="https://github.com/DaanV2/mechanus"
        by="DaanV2"
      />
      <FooterIcon
        href="https://github.com/DaanV2/mechanus"
        ariaLabel="Mechanus on GitHub"
        class="text-gray-300 hover:text-white"><GithubSolid /></FooterIcon
      >
    </div>
  </div>
</Footer>
