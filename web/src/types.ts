export interface TerminalLine {
  id: number;
  text: string;
  className: string;
  isHTML?: boolean;
}

export interface TerminalAPI {
  print: (text: string, className?: string) => void;
  printHTML: (html: string, className?: string) => void;
  clear: () => void;
  waitForInput: (prompt: string, isPassword?: boolean) => Promise<string>;
}
