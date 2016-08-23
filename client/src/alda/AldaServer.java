package alda;

import java.io.File;
import java.io.IOException;
import java.net.URISyntaxException;

import org.apache.commons.lang3.SystemUtils;

import org.fusesource.jansi.AnsiConsole;
import static org.fusesource.jansi.Ansi.*;
import static org.fusesource.jansi.Ansi.Color.*;

public class AldaServer extends AldaProcess {
  private static int PING_TIMEOUT = 100; // ms
  private static int PING_RETRIES = 5;
  private static int STARTUP_RETRY_INTERVAL = 250; // ms

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
    Color color = isGood ? GREEN : RED;
    String glyph = isGood ? CHECKMARK : X;
    msg(ansi().a("Server down ").fg(color).a(glyph).reset().toString());
  }

  private void serverDown() {
    serverDown(false);
  }

  public void upBg(int numberOfWorkers) throws InvalidOptionsException {
    assertNotRemoteHost();

    boolean serverAlreadyUp = checkForConnection();
    if (serverAlreadyUp) {
      msg("Server already up.");
      System.exit(1);
    }

    boolean serverAlreadyTryingToStart;
    try {
      serverAlreadyTryingToStart = SystemUtils.IS_OS_UNIX &&
                                   AldaClient.checkForExistingServer(this.port);
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
                     "--workers", Integer.toString(numberOfWorkers),
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

  public void upFg(int numberOfWorkers) throws InvalidOptionsException {
    assertNotRemoteHost();

    Object[] args = {numberOfWorkers, port};

    Util.callClojureFn("alda.server/start-server!", args);
  }

  // TODO: rewrite REPL as a client that communicates with a server
  public void startRepl() throws InvalidOptionsException {
    assertNotRemoteHost();

    Object[] args = {};

    Util.callClojureFn("alda.repl/start-repl!", args);
  }

  public void down() throws NoResponseException {
    boolean serverAlreadyDown = !checkForConnection();
    if (serverAlreadyDown) {
      msg("Server already down.");
      return;
    }

    msg("Stopping Alda server...");

    AldaRequest req = new AldaRequest(this.host, this.port);
    req.command = "stop-server";

    try {
      AldaResponse res = req.send();
    } catch (NoResponseException e) {
      serverDown(true);
      return;
    }

    boolean serverIsDown = waitForLackOfConnection();
    if (serverIsDown) {
      serverDown(true);
    } else {
      throw new NoResponseException("Failed to stop server.");
    }
  }

  public void downUp(int numberOfWorkers)
    throws NoResponseException, InvalidOptionsException {
    down();
    System.out.println();
    upBg(numberOfWorkers);
  }

  public void status() {
    boolean serverIsUp = checkForConnection();
    if (serverIsUp) {
      serverUp();
    } else {
      serverDown();
    }
  }

  public void version() throws NoResponseException {
    AldaRequest req = new AldaRequest(this.host, this.port);
    req.command = "version";
    AldaResponse res = req.send();
    String serverVersion = res.body;

    msg(serverVersion);
  }

  public void play(String code, String from, String to)
    throws NoResponseException {

    AldaRequest req = new AldaRequest(this.host, this.port);
    req.command = "play";
    req.body = code;
    req.options = new AldaRequestOptions();
    if (from != null) {
      req.options.from = from;
    }
    if (to != null) {
      req.options.to = to;
    }
    // play requests need to be sent exactly once and not retried, otherwise
    // the score could be played more than once.
    //
    // FIXME - implement "worker status" system so the client can poll to see
    // if the worker is handling its request
    AldaResponse res = req.send(3000, 0);

    if (res.success) {
      msg(res.body);
    } else {
      error(res.body);
    }
  }

  public void play(File file, String from, String to)
    throws NoResponseException {
    try {
      String fileBody = Util.readFile(file);
      play(fileBody, from, to);
    } catch (IOException e) {
      error("Unable to read file: " + file.getAbsolutePath());
    }
  }

  public void parse(String code, String mode) throws NoResponseException {
    AldaRequest req = new AldaRequest(this.host, this.port);
    req.command = "parse";
    req.body = code;
    req.options = new AldaRequestOptions();
    req.options.as = mode;
    AldaResponse res = req.send();

    if (res.success) {
      System.out.println(res.body);
    } else {
      error(res.body);
    }
  }

  public void parse(File file, String mode) throws NoResponseException {
    try {
      String fileBody = Util.readFile(file);
      parse(fileBody, mode);
    } catch (IOException e) {
      error("Unable to read file: " + file.getAbsolutePath());
    }
  }
}
