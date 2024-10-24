<script>
	import { twMerge } from 'tailwind-merge';
</script>

<svg
	xmlns="http://www.w3.org/2000/svg"
	class={twMerge(
		'h-5 w-5 flex-shrink-0 cursor-pointer text-gray-500 hover:text-red-500',
		$$props.class
	)}
	viewBox="0 0 24 24"
	fill="none"
	stroke="currentColor"
	stroke-width="2"
	stroke-linecap="round"
	stroke-linejoin="round"
>
	<path d="M3 6h18"></path>
	<path d="M8 6V4h8v2"></path>
	<path d="M19 6l-1.5 13.1A2 2 0 0 1 15.5 21H8.5A2 2 0 0 1 6.5 19.1L5 6"></path>
	<path d="M10 11v6"></path>
	<path d="M14 11v6"></path>
</svg>
