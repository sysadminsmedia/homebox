import { exec } from "child_process";

function globalTeardown() {
  if (process.env.TEST_SHUTDOWN_API_SERVER) {
    const pc = exec("pkill -SIGTERM api"); // Kill background API process
    const fr = exec("pkill -SIGTERM task"); // Kill background Frontend process
    pc.stdout?.on("data", (data: void) => {
      console.log(`stdout: ${data}`);
    });
    fr.stdout?.on("data", (data: void) => {
      console.log(`stdout: ${data}`);
    });
  }
}

export default globalTeardown;
