<script lang="ts">
  import type { ManifestFile } from '../../lib/api/types';

  interface Props {
    files: ManifestFile[];
    selectedFile: string;
    onselect: (path: string) => void;
  }

  let { files, selectedFile, onselect }: Props = $props();

  interface TreeNode {
    name: string;
    path: string;
    children: TreeNode[];
    isFile: boolean;
  }

  let tree = $derived(buildTree(files));

  function buildTree(fileList: ManifestFile[]): TreeNode[] {
    const root: TreeNode[] = [];
    const paths = fileList.map((f) => f.path).sort();

    for (const path of paths) {
      const parts = path.split('/');
      let current = root;

      for (let i = 0; i < parts.length; i++) {
        const name = parts[i];
        const fullPath = parts.slice(0, i + 1).join('/');
        const isFile = i === parts.length - 1;
        let node = current.find((n) => n.name === name);

        if (!node) {
          node = { name, path: fullPath, children: [], isFile };
          current.push(node);
        }
        current = node.children;
      }
    }

    return root;
  }

  let expanded = $state<Set<string>>(new Set());

  function toggle(path: string) {
    const next = new Set(expanded);
    if (next.has(path)) next.delete(path);
    else next.add(path);
    expanded = next;
  }
</script>

<div class="file-explorer">
  <div class="explorer-header">Files</div>
  <div class="explorer-tree">
    {#each tree as node}
      {@render treeNode(node, 0)}
    {/each}
  </div>
</div>

{#snippet treeNode(node: TreeNode, depth: number)}
  {#if node.isFile}
    <button
      class="tree-item file"
      class:selected={selectedFile === node.path}
      style="padding-left: {12 + depth * 16}px"
      onclick={() => onselect(node.path)}
    >
      <span class="tree-icon file-icon"></span>
      {node.name}
    </button>
  {:else}
    <button
      class="tree-item folder"
      style="padding-left: {12 + depth * 16}px"
      onclick={() => toggle(node.path)}
    >
      <span class="tree-icon">{expanded.has(node.path) ? '\u25BE' : '\u25B8'}</span>
      {node.name}
    </button>
    {#if expanded.has(node.path)}
      {#each node.children as child}
        {@render treeNode(child, depth + 1)}
      {/each}
    {/if}
  {/if}
{/snippet}

<style>
  .file-explorer {
    display: flex;
    flex-direction: column;
    height: 100%;
    background: var(--bg-elevated);
    border: 1px solid var(--border-subtle);
    border-radius: var(--radius-md);
    overflow: hidden;
  }

  .explorer-header {
    padding: var(--space-sm) var(--space-md);
    font-size: 0.6875rem;
    font-weight: 600;
    text-transform: uppercase;
    letter-spacing: 0.05em;
    color: var(--text-tertiary);
    border-bottom: 1px solid var(--border-subtle);
  }

  .explorer-tree {
    flex: 1;
    overflow-y: auto;
    padding: var(--space-xs) 0;
  }

  .tree-item {
    display: flex;
    align-items: center;
    gap: 6px;
    width: 100%;
    padding: 4px 12px;
    font-size: 0.8125rem;
    color: var(--text-secondary);
    text-align: left;
    transition: background var(--duration-fast) var(--ease-out);
  }

  .tree-item:hover {
    background: var(--bg-surface);
    color: var(--text-primary);
  }

  .tree-item.selected {
    background: var(--accent-subtle);
    color: var(--accent);
  }

  .tree-icon {
    font-size: 0.75rem;
    flex-shrink: 0;
    width: 14px;
    text-align: center;
  }

  .file-icon::before {
    content: '\2022';
    color: var(--text-tertiary);
  }
</style>
