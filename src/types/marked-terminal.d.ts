declare module 'marked-terminal' {
  import { MarkedExtension } from 'marked';

  interface TerminalRendererOptions {
    codespan?: (text: string) => string;
    code?: (code: string, language: string | undefined) => string;
    showSectionPrefix?: boolean;
    tab?: number;
    [key: string]: unknown;
  }

  export default class TerminalRenderer implements MarkedExtension {
    constructor(options?: TerminalRendererOptions);
  }
}
