import type { TerminalAPI } from './types';
import { rdbAPI, kvAPI, workerAPI, domainAPI, getAuthState } from './api';

function requireAuth(terminal: TerminalAPI): boolean {
  if (!getAuthState().token) {
    terminal.print('Please login first', 'error');
    return false;
  }
  return true;
}

function formatBytes(bytes: number): string {
  if (bytes <= 0) return '0 B';
  const units = ['B', 'KB', 'MB', 'GB'];
  const i = Math.min(Math.floor(Math.log(bytes) / Math.log(1024)), units.length - 1);
  return `${(bytes / Math.pow(1024, i)).toFixed(2)} ${units[i]}`;
}

// === RDB Commands ===

async function rdbList(terminal: TerminalAPI) {
  try {
    const result = await rdbAPI.list();
    terminal.print('');
    terminal.print('=== RDB Resources ===', 'info');
    if (result.database_size !== undefined) {
      terminal.print(`Database Total: ${formatBytes(result.database_size)}`, 'info');
    }
    terminal.print('');
    if (result.rdbs && result.rdbs.length > 0) {
      result.rdbs.forEach((rdb: { id: string; name: string; url: string; size: number }) => {
        terminal.print(`ID: ${rdb.id}`, 'success');
        terminal.print(`  Name: ${rdb.name}`);
        terminal.print(`  URL: ${rdb.url}`);
        terminal.print(`  Size: ${formatBytes(rdb.size)}`);
        terminal.print('');
      });
    } else {
      terminal.print('No RDB resources found', 'warning');
    }
  } catch (error) {
    terminal.print(`Failed to list RDBs: ${(error as Error).message}`, 'error');
  }
}

async function rdbAdd(terminal: TerminalAPI) {
  try {
    const name = await terminal.waitForInput('Enter RDB name:');
    const result = await rdbAPI.create(name);
    terminal.print('', 'success');
    terminal.print(`ID: ${result.id}`, 'info');
    terminal.print(result.message, 'success');
  } catch (error) {
    terminal.print(`Failed to create RDB: ${(error as Error).message}`, 'error');
  }
}

async function rdbDelete(terminal: TerminalAPI, id: string) {
  try {
    await rdbAPI.delete(id);
    terminal.print('RDB deleted successfully', 'success');
  } catch (error) {
    terminal.print(`Failed to delete RDB: ${(error as Error).message}`, 'error');
  }
}

export async function rdbCommand(terminal: TerminalAPI, args: string[]) {
  if (!requireAuth(terminal)) return;
  switch (args[0]) {
    case 'list': await rdbList(terminal); break;
    case 'add': await rdbAdd(terminal); break;
    case 'delete':
      if (!args[1]) { terminal.print('Usage: rdb delete <id>', 'error'); return; }
      await rdbDelete(terminal, args[1]); break;
    default: terminal.print('Usage: rdb [list|add|delete]', 'error');
  }
}

// === KV Commands ===

async function kvList(terminal: TerminalAPI) {
  try {
    const result = await kvAPI.list();
    terminal.print('');
    terminal.print('=== KV Resources ===', 'info');
    if (result.kvs && result.kvs.length > 0) {
      result.kvs.forEach((kv: { id: string; kv_type: string; url: string }) => {
        terminal.print(`ID: ${kv.id}`, 'success');
        terminal.print(`  Type: ${kv.kv_type}`);
        terminal.print(`  URL: ${kv.url}`);
        terminal.print('');
      });
    } else {
      terminal.print('No KV resources found', 'warning');
    }
  } catch (error) {
    terminal.print(`Failed to list KVs: ${(error as Error).message}`, 'error');
  }
}

async function kvAdd(terminal: TerminalAPI) {
  try {
    const type = await terminal.waitForInput('Enter KV type (redis/memory):');
    const url = await terminal.waitForInput('Enter KV URL:');
    const result = await kvAPI.create(type, url);
    terminal.print('', 'success');
    terminal.print(`ID: ${result.id}`, 'info');
    if (result.error) {
      terminal.print(result.error, 'warning');
    } else {
      terminal.print(result.message, 'success');
    }
  } catch (error) {
    terminal.print(`Failed to create KV: ${(error as Error).message}`, 'error');
  }
}

async function kvDelete(terminal: TerminalAPI, id: string) {
  try {
    await kvAPI.delete(id);
    terminal.print('KV deleted successfully', 'success');
  } catch (error) {
    terminal.print(`Failed to delete KV: ${(error as Error).message}`, 'error');
  }
}

export async function kvCommand(terminal: TerminalAPI, args: string[]) {
  if (!requireAuth(terminal)) return;
  switch (args[0]) {
    case 'list': await kvList(terminal); break;
    case 'add': await kvAdd(terminal); break;
    case 'delete':
      if (!args[1]) { terminal.print('Usage: kv delete <id>', 'error'); return; }
      await kvDelete(terminal, args[1]); break;
    default: terminal.print('Usage: kv [list|add|delete]', 'error');
  }
}

// === Worker Commands ===

async function workerList(terminal: TerminalAPI) {
  try {
    const result = await workerAPI.list();
    terminal.print('');
    terminal.print('=== Workers ===', 'info');
    if (result && result.length > 0) {
      result.forEach((w: { worker_id: string; image: string; port: number }) => {
        terminal.print(`ID: ${w.worker_id}`, 'success');
        terminal.print(`  Image: ${w.image}`);
        terminal.print(`  Port: ${w.port}`);
        terminal.print('');
      });
    } else {
      terminal.print('No workers found', 'warning');
    }
  } catch (error) {
    terminal.print(`Failed to list workers: ${(error as Error).message}`, 'error');
  }
}

async function workerAdd(terminal: TerminalAPI) {
  try {
    const worker_id = await terminal.waitForInput('Enter worker ID:');
    const image = await terminal.waitForInput('Enter Docker image:');
    const port = await terminal.waitForInput('Enter port:');
    const result = await workerAPI.register(worker_id, image, parseInt(port));
    terminal.print('', 'success');
    terminal.print(`Worker ID: ${result.worker_id}`, 'info');
    terminal.print(`Status: ${result.status}`, result.status === 'active' ? 'success' : 'warning');
    if (result.domain) {
      terminal.print(`Domain: ${result.domain}`, 'info');
    }
    if (result.error) {
      terminal.print(`Error: ${result.error}`, 'error');
    }
  } catch (error) {
    terminal.print(`Failed to create worker: ${(error as Error).message}`, 'error');
  }
}

async function workerDelete(terminal: TerminalAPI, id: string) {
  try {
    await workerAPI.delete(id);
    terminal.print('Worker deleted successfully', 'success');
  } catch (error) {
    terminal.print(`Failed to delete worker: ${(error as Error).message}`, 'error');
  }
}

export async function workerCommand(terminal: TerminalAPI, args: string[]) {
  if (!requireAuth(terminal)) return;
  switch (args[0]) {
    case 'list': await workerList(terminal); break;
    case 'add': await workerAdd(terminal); break;
    case 'delete':
      if (!args[1]) { terminal.print('Usage: worker delete <id>', 'error'); return; }
      await workerDelete(terminal, args[1]); break;
    default: terminal.print('Usage: worker [list|add|delete]', 'error');
  }
}

// === Domain Commands ===

async function domainList(terminal: TerminalAPI) {
  try {
    const result = await domainAPI.list();
    terminal.print('');
    terminal.print('=== Custom Domains ===', 'info');
    if (result.domains && result.domains.length > 0) {
      result.domains.forEach((d: { id: string; domain: string; target: string; status: string }) => {
        terminal.print(`ID: ${d.id}`, 'success');
        terminal.print(`  Domain: ${d.domain}`);
        terminal.print(`  Target: ${d.target}`);
        terminal.print(`  Status: ${d.status}`);
        terminal.print('');
      });
    } else {
      terminal.print('No custom domains found', 'warning');
    }
  } catch (error) {
    terminal.print(`Failed to list domains: ${(error as Error).message}`, 'error');
  }
}

async function domainAdd(terminal: TerminalAPI) {
  try {
    const domain = await terminal.waitForInput('Enter your domain (e.g. api.example.com):');
    const target = await terminal.waitForInput('Enter target domain to proxy to:');
    const result = await domainAPI.create(domain, target);
    terminal.print('', 'success');
    terminal.print('Domain verification required!', 'warning');
    terminal.print('');
    terminal.print('Add this TXT record to your DNS:', 'info');
    terminal.print(`  Name:  ${result.txt_name}`);
    terminal.print(`  Value: ${result.txt_value}`);
    terminal.print('');
    terminal.print(`ID: ${result.id}`, 'info');
    terminal.print(`Status: ${result.status}`, 'warning');
    terminal.print('');
    terminal.print('Verification will run for 60 seconds...', 'info');
  } catch (error) {
    terminal.print(`Failed to add domain: ${(error as Error).message}`, 'error');
  }
}

async function domainGet(terminal: TerminalAPI, id: string) {
  try {
    const result = await domainAPI.get(id);
    terminal.print('');
    terminal.print(`ID: ${result.id}`, 'success');
    terminal.print(`  Domain: ${result.domain}`);
    terminal.print(`  Target: ${result.target}`);
    terminal.print(`  Status: ${result.status}`);
    terminal.print(`  TXT Name: ${result.txt_name}`);
    terminal.print(`  TXT Value: ${result.txt_value}`);
  } catch (error) {
    terminal.print(`Failed to get domain: ${(error as Error).message}`, 'error');
  }
}

async function domainDelete(terminal: TerminalAPI, id: string) {
  try {
    await domainAPI.delete(id);
    terminal.print('Domain deleted successfully', 'success');
  } catch (error) {
    terminal.print(`Failed to delete domain: ${(error as Error).message}`, 'error');
  }
}

export async function domainCommand(terminal: TerminalAPI, args: string[]) {
  if (!requireAuth(terminal)) return;
  switch (args[0]) {
    case 'list': await domainList(terminal); break;
    case 'add': await domainAdd(terminal); break;
    case 'get':
      if (!args[1]) { terminal.print('Usage: domain get <id>', 'error'); return; }
      await domainGet(terminal, args[1]); break;
    case 'delete':
      if (!args[1]) { terminal.print('Usage: domain delete <id>', 'error'); return; }
      await domainDelete(terminal, args[1]); break;
    default: terminal.print('Usage: domain [list|add|get|delete]', 'error');
  }
}
