export class Mutex {
  private locked: boolean;

  constructor() {
    this.locked = false;
  }

  async lock(): Promise<void> {
    while (this.locked) {
      const timeout_ms = 100;
      await new Promise((resolve) => setTimeout(resolve, timeout_ms));
    }
    this.locked = true;
  }

  unlock() {
    this.locked = false;
  }
}
