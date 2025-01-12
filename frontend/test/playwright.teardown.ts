import { exec } from "child_process";

// eslint-disable-next-line require-await
export async function teardown() {
  if (process.env.TEST_SHUTDOWN_API_SERVER) {
    const pc = exec("pkill -SIGTERM api"); // Kill background API process
    const fr = exec("pkill -SIGTERM node"); // Kill background Frontend process
    pc.stdout?.on("data", (data: void) => {
      console.log(`stdout: ${data}`);
    });
    fr.stdout?.on("data", (data: void) => {
      console.log(`stdout: ${data}`);
    });
  }
}
