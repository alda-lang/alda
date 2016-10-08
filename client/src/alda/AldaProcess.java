package alda;

import java.util.concurrent.Callable;
import java.util.concurrent.TimeUnit;

import net.jodah.recurrent.Recurrent;
import net.jodah.recurrent.RetryPolicy;

public class AldaProcess {
  private static int PING_TIMEOUT = 100; // ms
  private static int PING_RETRIES = 5;
  private static int STARTUP_RETRY_INTERVAL = 250; // ms

  public boolean verbose = false;
  public String host;
  public int pid;
  public int port;
  public String type;
  public int timeout;

  public boolean checkForConnection(int timeout, int retries) {
    try {
      AldaRequest req = new AldaRequest(this.host, this.port);
      req.command = "ping";
      AldaResponse res = req.send(timeout, retries);
      return res.success;
    } catch (NoResponseException e) {
      return false;
    }
  }

  public boolean checkForConnection() {
    return checkForConnection(PING_TIMEOUT, PING_RETRIES);
  }

  public boolean waitForConnection() {
    // Calculate the number of retries before giving up, based on the fixed
    // STARTUP_RETRY_INTERVAL and the desired timeout in seconds.
    int retriesPerSecond = 1000 / this.STARTUP_RETRY_INTERVAL;
    int retries = this.timeout * retriesPerSecond;

    return checkForConnection(this.STARTUP_RETRY_INTERVAL, retries);
  }

  public boolean waitForLackOfConnection() {
    RetryPolicy retryPolicy = new RetryPolicy()
      .withDelay(500, TimeUnit.MILLISECONDS)
      .withMaxDuration(30, TimeUnit.SECONDS)
      .retryFor(false);

    Callable<Boolean> ping = new Callable<Boolean>() {
      public Boolean call() { return !checkForConnection(); }
    };

    return Recurrent.get(ping, retryPolicy);
  }

}
