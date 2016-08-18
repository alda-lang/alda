package alda;

import java.io.File;
import java.io.IOException;
import java.net.URISyntaxException;
import java.util.concurrent.Callable;
import java.util.concurrent.TimeUnit;

import org.fusesource.jansi.AnsiConsole;
import static org.fusesource.jansi.Ansi.*;
import static org.fusesource.jansi.Ansi.Color.*;

import net.jodah.recurrent.Recurrent;
import net.jodah.recurrent.RetryPolicy;

public class AldaServer {
  private static int PING_TIMEOUT = 100;    // ms
  private static int PING_RETRIES = 5;
  private static int STARTUP_TIMEOUT = 250; // ms
  private static int STARTUP_RETRIES = 60;

  private String host;
  private int port;

  public AldaServer(String host, int port) {
    this.host = normalizeHost(host);
    this.port = port;

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
    return checkForConnection(STARTUP_TIMEOUT, STARTUP_RETRIES);
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

  public void startBg() throws InvalidOptionsException {
    assertNotRemoteHost();

    boolean serverAlreadyUp = checkForConnection();
    if (serverAlreadyUp) {
      msg("Server already up.");
      return;
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

  public void startFg() throws InvalidOptionsException {
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

  public void stop(boolean autoConfirm) throws ServerResponseException {
    boolean serverAlreadyDown = !checkForConnection();
    if (serverAlreadyDown) {
      msg("Server already down.");
      return;
    }

    msg("Stopping Alda server...");

    AldaServerRequest req = new AldaServerRequest(this.host, this.port);
    req.command = "stop-server";

    if (autoConfirm) {
      req.confirming = true;
    }

    try {
      AldaServerResponse res = req.send();

      if (res.signal != null && res.signal.equals("unsaved-changes")) {
        System.out.println();

        boolean confirm =
          Util.promptForConfirmation("The score has unsaved changes that will " +
              "be lost.\nAre you sure you want to stop " +
              "the server?");
        if (confirm) {
          System.out.println();
          stop(true);
        }

        return;
      }
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

  public void restart(boolean autoConfirm)
    throws ServerResponseException, InvalidOptionsException {
    stop(autoConfirm);
    System.out.println();
    startBg();
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

  public String getFilename() throws ServerResponseException {
    AldaServerRequest req = new AldaServerRequest(this.host, this.port);
    req.command = "filename";
    AldaServerResponse res = req.send();

    return res.body;
  }

  public boolean isScoreModified() throws ServerResponseException {
    AldaServerRequest req = new AldaServerRequest(this.host, this.port);
    req.command = "modified?";
    AldaServerResponse res = req.send();

    return (res.body.equals("true"));
  }

  public void info() throws ServerResponseException {
    AldaServerRequest req = new AldaServerRequest(this.host, this.port);
    req.command = "info";
    AldaServerResponse res = req.send();

    if (res.success) {
      System.out.println(res.body);
    } else {
      error(res.body);
    }
  }

  public void score(String mode) throws ServerResponseException {
    AldaServerRequest req = new AldaServerRequest(this.host, this.port);
    req.command = "current-score";
    req.options = new AldaServerRequestOptions();
    req.options.as = mode;
    AldaServerResponse res = req.send();

    System.out.println(res.body);
  }

  public void delete(boolean autoConfirm) throws ServerResponseException {
    AldaServerRequest req = new AldaServerRequest(this.host, this.port);
    req.command = "new-score";
    req.confirming = autoConfirm;
    AldaServerResponse res = req.send();

    if (res.signal != null && res.signal.equals("unsaved-changes")) {
      System.out.println();
      boolean confirm =
        Util.promptForConfirmation("The current score has unsaved changes " +
                                   "that will be lost.\nAre you sure you " +
                                   "want to start a new score?");
      if (confirm) {
        delete(true);
      }

      return;
    }

    msg("New score initialized.");
  }

  public void load(String code, String filename, boolean autoConfirm)
    throws ServerResponseException {
    AldaServerRequest req = new AldaServerRequest(this.host, this.port);
    req.command = "load";
    req.body = code;
    req.confirming = autoConfirm;
    if (filename != null) {
      req.options = new AldaServerRequestOptions();
      req.options.filename = filename;
    }
    AldaServerResponse res = req.send();

    if (res.signal != null && res.signal.equals("unsaved-changes")) {
      System.out.println();
      boolean confirm =
        Util.promptForConfirmation("The current score has unsaved changes " +
                                   "that will be lost.\nAre you sure you " +
                                   "want to proceed?");
      if (confirm) {
        load(code, filename, true);
      }

      return;
    }

    msg(filename != null ? "Loaded file." : "Loaded code.");
  }

  public void load(String code, boolean autoConfirm)
    throws ServerResponseException {
    load(code, null, autoConfirm);
  }

  public void load(File file, boolean autoConfirm)
    throws ServerResponseException {
    try {
      String fileBody = Util.readFile(file);
      String filename = file.getAbsolutePath();
      load(fileBody, filename, autoConfirm);
    } catch (IOException e) {
      error("Unable to read file: " + file.getAbsolutePath());
    }
  }

  private void saveImpl(String filename, boolean autoConfirm)
    throws ServerResponseException {
    AldaServerRequest req = new AldaServerRequest(this.host, this.port);
    req.command = "save";
    req.confirming = autoConfirm;
    if (filename != null) {
      req.options = new AldaServerRequestOptions();
      req.options.filename = filename;
    }
    AldaServerResponse res = req.send();

    if (res.signal != null && res.signal.equals("existing-file")) {
      System.out.println();
      boolean confirm =
        Util.promptForConfirmation("There is an existing file with the " +
                                   "filename you specified. Saving the score " +
                                   "to this file will erase whatever is " +
                                   "already there.\n\n" +
                                   "Are you sure you want to do this?");
      if (confirm) {
        saveImpl(filename, true);
      }

      return;
    }

    if (res.success) {
      msg(res.body);
    } else {
      throw new ServerRuntimeError(res.body);
    }
  }

  public void save(File file, boolean autoConfirm)
    throws ServerResponseException {
    String filename = file.getAbsolutePath();
    saveImpl(filename, autoConfirm);
  }

  public void save() throws ServerResponseException {
    saveImpl(null, false);
  }

  public void play(String from, String to)
    throws ServerResponseException {
    AldaServerRequest req = new AldaServerRequest(this.host, this.port);
    req.command = "play-score";
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

  public void play(String code, String from, String to, boolean appendToScore)
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
    req.options.append = appendToScore;
    AldaServerResponse res = req.send();

    if (res.success) {
      msg(res.body);
    } else {
      error(res.body);
    }
  }

  public void play(File file, String from, String to, boolean appendToScore)
    throws ServerResponseException {
    try {
      String fileBody = Util.readFile(file);
      play(fileBody, from, to, appendToScore);
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

  public void append(String code) throws ServerResponseException {
    AldaServerRequest req = new AldaServerRequest(this.host, this.port);
    req.command = "append";
    req.body = code;
    AldaServerResponse res = req.send();

    if (res.success) {
      msg("Appended code to score.");
    } else {
      error(res.body);
    }
  }

  public void append(File file) throws ServerResponseException {
    try {
      String fileBody = Util.readFile(file);
      append(fileBody);
    } catch (IOException e) {
      error("Unable to read file: " + file.getAbsolutePath());
    }
  }
}
