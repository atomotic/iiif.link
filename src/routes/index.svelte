<script>
	import { onMount } from 'svelte/internal';
	import * as manifesto from 'manifesto.js';
	import OpenSeadragon from 'openseadragon?client';
	import Annotorious from '@recogito/annotorious-openseadragon?client';
	import Toolbar from '@recogito/annotorious-toolbar?client';
	import '@recogito/annotorious-openseadragon/dist/annotorious.min.css';

	let promise;
	let osd;
	let anno;

	let manifest =
		'https://fondospaccini.dhmore.jarvis.memooria.org/meta/iiif/958c1078-fb69-40d0-8a3f-b1d2d3058680/manifest';

	async function fetchManifest() {
		const loadManifest = await manifesto.loadManifest(manifest);
		const parsedManifest = manifesto.parseManifest(loadManifest);
		return parsedManifest;
	}

	function loadManifest() {
		promise = fetchManifest();
	}

	onMount(() => {
		osd = OpenSeadragon({
			id: 'viewer',
			tileSources: []
		});

		anno = Annotorious(osd);
		anno.disableEditor = true;
		anno.removeDrawingTool("polygon");
		Toolbar(anno, document.getElementById('toolbar'));

		return () => {
			anno.destroy();
			osd.destroy();
		};
	});
</script>

<div id="form">
	<input type="text" size="80" bind:value={manifest} />
	<button on:click={loadManifest}>open IIIF manifest</button>
</div>

<div id="browser">
	{#if promise != null}
		{#await promise}
			<p />
			loading
		{:then data}
			{#each data.getSequenceByIndex(0).getCanvases() as canvas}
				<img
					src="{canvas.getImages()[0].getResource().id}/full/,90/0/default.jpg"
					alt=""
					on:click={() => osd.open(`${canvas.getImages()[0].getResource().id}/info.json`)}
				/>
			{/each}
		{:catch error}
			<p>An error occurred!</p>
		{/await}
	{/if}
</div>
<div id="toolbar"></div>
<div id="viewer" style="width: calc(100vw-20px); height: 80vh" />

<style>
	:root {
		font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Oxygen, Ubuntu, Cantarell,
			'Open Sans', 'Helvetica Neue', sans-serif;
	}
	#browser {
		padding: 5px;
		min-height: 14vh;
		display: flex;
		overflow-x: auto;
		overflow-y: hidden;
		align-items: center;
	}
	#browser img {
		margin-right: 15px;
		display: block;
	}
</style>
