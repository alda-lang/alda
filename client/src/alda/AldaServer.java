package alda;

import java.io.File;
import java.io.IOException;
import java.net.URISyntaxException;
import java.util.concurrent.Callable;
import java.util.concurrent.TimeUnit;

import org.apache.commons.lang3.SystemUtils;

import org.fusesource.jansi.AnsiConsole;
import static org.fusesource.jansi.Ansi.*;
import static org.fusesource.jansi.Ansi.Color.*;

import net.jodah.recurrent.Recurrent;
import net.jodah.recurrent.RetryPolicy;

public class AldaServer {
  private static int PING_TIMEOUT = 100; // ms
  private static int PING_RETRIES = 5;
  private static int STARTUP_RETRY_INTERVAL = 250; // ms

  private String host;
  private int port;
  private int timeout;

  public AldaServer(String host, int port, int timeout) {
    this.host = normalizeHost(host);
    this.port = port;
    this.timeout = timeout;

    AnsiConsole.systemInstall();
  }

  private static String normalizeHost(String host) {
    // trim leading/trailing whitespace and trailing "/"
    host = host.trim().replaceAll("/$", "");
    // prepend tcp:// if not already present
    if (!(host.startsWith("tcp://"))) {
      host = "tcp://" + host;
    }
    return host;
  }

  private void assertNotRemoteHost() throws InvalidOptionsException {
    String hostWithoutProtocol = host.replaceAll("tcp://", "");

    if (!hostWithoutProtocol.equals("localhost")) {
      throw new InvalidOptionsException(
          "Alda servers cannot be started remotely.");
    }
  }

  public void msg(String message, Object... args) {
    String hostWithoutProtocol = host.replaceAll("tcp://", "");

    String prefix;
    if (hostWithoutProtocol.equals("localhost")) {
      prefix = "";
    } else {
      prefix = hostWithoutProtocol + ":";
    }

    prefix += Integer.toString(port);
    prefix = String.format("[%s] ", ansi().fg(BLUE)
                                          .a(prefix)
                                          .reset()
                                          .toString());

    System.out.printf(prefix + message + "\n", args);
  }

  public void error(String message, Object... args) {
    String prefix = ansi().fg(RED).a("ERROR ").reset().toString();
    msg(prefix + message, args);
  }

  private final String CHECKMARK = "\u2713";
  private final String X = "\u2717";

  private void serverUp() {
    msg(ansi().a("Server up ").fg(GREEN).a(CHECKMARK).reset().toString());
  }

  private void serverDown(boolean isGood) {
    Color color = isGood? GREEN : RED;
    String glyph = isGood ? CHECKMARK : X;
    msg(ansi().a("Server down ").fg(color).a(glyph).reset().toString());
  }

  private void serverDown() {
    serverDown(false);
  }

  private boolean checkForConnection(int timeout, int retries) {
    try {
      AldaServerRequest req = new AldaServerRequest(this.host, this.port);
      req.command = "ping";
      AldaServerResponse res = req.send(timeout, retries);
      return res.success;
    } catch (ServerResponseException e) {
      return false;
    }
  }

  private boolean checkForConnection() {
    return checkForConnection(PING_TIMEOUT, PING_RETRIES);
  }

  private boolean waitForConnection() {
    // Calculate the number of retries before giving up, based on the fixed
    // STARTUP_RETRY_INTERVAL and the desired timeout in seconds.
    int retriesPerSecond = 1000 / STARTUP_RETRY_INTERVAL;
    int retries = this.timeout * retriesPerSecond;

    return checkForConnection(STARTUP_RETRY_INTERVAL, retries);
  }

  private boolean waitForLackOfConnection() {
    RetryPolicy retryPolicy = new RetryPolicy()
      .withDelay(500, TimeUnit.MILLISECONDS)
      .withMaxDuration(30, TimeUnit.SECONDS)
      .retryFor(false);

    Callable<Boolean> ping = new Callable<Boolean>() {
      public Boolean call() { return !checkForConnection(); }
    };

    return Recurrent.get(ping, retryPolicy);
  }

  public void upBg() throws InvalidOptionsException {
    assertNotRemoteHost();

    boolean serverAlreadyUp = checkForConnection();
    if (serverAlreadyUp) {
      msg("Server already up.");
      return;
    }

    boolean serverAlreadyTryingToStart;
    try {
      serverAlreadyTryingToStart = SystemUtils.IS_OS_UNIX &&
                                   Util.checkForExistingServer(this.port);
    } catch (IOException e) {
      System.out.println("WARNING: Unable to detect whether or not there is " +
                         "already a server running on that port.");
      serverAlreadyTryingToStart = false;
    }

    if (serverAlreadyTryingToStart) {
      msg("There is already a server trying to start on this port. Please " +
          "be patient -- this can take a while.");
      System.exit(1);
    }

    Object[] opts = {"--host", host,
                     "--port", Integer.toString(port),
                     "--alda-fingerprint"};

    try {
      Util.forkProgram(Util.conj(opts, "server"));
      msg("Starting Alda server...");

      boolean serverUp = waitForConnection();
      if (serverUp) {
        serverUp();
      } else {
        serverDown();
      }
    } catch (URISyntaxException e) {
      error("Unable to fork '%s' into the background; " +
            " got URISyntaxException: %s", e.getInput(), e.getReason());
    } catch (IOException e) {
      error("An IOException occurred trying to fork a background process: %s",
            e.getMessage());
    }
  }

  public void upFg() throws InvalidOptionsException {
    assertNotRemoteHost();

    Object[] args = {port};

    Util.callClojureFn("alda.server/start-server!", args);
  }

  // TODO: rewrite REPL as a client that communicates with a server
  public void startRepl() throws InvalidOptionsException {
    assertNotRemoteHost();

    Object[] args = {};

    Util.callClojureFn("alda.repl/start-repl!", args);
  }

  public void down() throws ServerResponseException {
    boolean serverAlreadyDown = !checkForConnection();
    if (serverAlreadyDown) {
      msg("Server already down.");
      return;
    }

    msg("Stopping Alda server...");

    AldaServerRequest req = new AldaServerRequest(this.host, this.port);
    req.command = "stop-server";

    try {
      AldaServerResponse res = req.send();
    } catch (ServerResponseException e) {
      serverDown(true);
      return;
    }

    boolean serverIsDown = waitForLackOfConnection();
    if (serverIsDown) {
      serverDown(true);
    } else {
      throw new ServerResponseException("Failed to stop server.");
    }
  }

  public void downUp()
    throws ServerResponseException, InvalidOptionsException {
    down();
    System.out.println();
    upBg();
  }

  public void status() {
    boolean serverIsUp = checkForConnection();
    if (serverIsUp) {
      serverUp();
    } else {
      serverDown();
    }
  }

  public void version() throws ServerResponseException {
    AldaServerRequest req = new AldaServerRequest(this.host, this.port);
    req.command = "version";
    AldaServerResponse res = req.send();
    String serverVersion = res.body;

    msg(serverVersion);
  }

  public void play(String code, String from, String to)
    throws ServerResponseException {

    AldaServerRequest req = new AldaServerRequest(this.host, this.port);
    req.command = "play";
    req.body = code;
    req.options = new AldaServerRequestOptions();
    if (from != null) {
      req.options.from = from;
    }
    if (to != null) {
      req.options.to = to;
    }
    AldaServerResponse res = req.send();

    if (res.success) {
      msg(res.body);
    } else {
      error(res.body);
    }
  }

  public void play(File file, String from, String to)
    throws ServerResponseException {
    try {
      String fileBody = Util.readFile(file);
      play(fileBody, from, to);
    } catch (IOException e) {
      error("Unable to read file: " + file.getAbsolutePath());
    }
  }

  public void parse(String code, String mode) throws ServerResponseException {
    AldaServerRequest req = new AldaServerRequest(this.host, this.port);
    req.command = "parse";
    req.body = code;
    req.options = new AldaServerRequestOptions();
    req.options.as = mode;
    AldaServerResponse res = req.send();

    if (res.success) {
      System.out.println(res.body);
    } else {
      error(res.body);
    }
  }

  public void parse(File file, String mode) throws ServerResponseException {
    try {
      String fileBody = Util.readFile(file);
      parse(fileBody, mode);
    } catch (IOException e) {
      error("Unable to read file: " + file.getAbsolutePath());
    }
  }
}
