package alda;

import org.fusesource.jansi.AnsiConsole;
import static org.fusesource.jansi.Ansi.*;
import static org.fusesource.jansi.Ansi.Color.*;

import clojure.lang.Keyword;

public class AldaServer {
  private String host;
  private int port;
  private int preBuffer;
  private int postBuffer;
  private boolean useStockSoundfont;

  public String getHost() { return host; }
  public int getPort() { return port; }

  public AldaServer(String host, int port, int preBuffer, int postBuffer,
                    boolean useStockSoundfont) {
    this.host = normalizeHost(host);
    this.port = port;
    this.preBuffer = preBuffer;
    this.postBuffer = postBuffer;
    this.useStockSoundfont = useStockSoundfont;
    AnsiConsole.systemInstall();
  }

  private static String normalizeHost(String host) {
    // trim leading/trailing whitespace and trailing "/"
    host = host.trim().replaceAll("/$", "");
    // prepend http:// if not already present
    if (!(host.startsWith("http://") || host.startsWith("https://"))) {
      host = "http://" + host;
    }
    return host;
  }


  private void validateNotRemoteHost() throws InvalidOptionsException {
    String hostWithoutProtocol = host.replaceAll("https?://", "");

    if (!hostWithoutProtocol.equals("localhost")) {
      throw new InvalidOptionsException("Alda servers cannot be started " +
          "remotely.");
    }
  }

  public void msg(String message, Object... args) {
    String hostWithoutProtocol = host.replaceAll("https?://", "");

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

  public void startBg()
    throws InvalidOptionsException, java.net.URISyntaxException, java.io.IOException {
    validateNotRemoteHost();

    Object[] opts = {"--host", host, "--port", Integer.toString(port),
                     "--pre-buffer", Integer.toString(preBuffer),
                     "--post-buffer", Integer.toString(postBuffer)};

    if (useStockSoundfont) {
      opts = Util.conj(opts, "--stock");
    }

    Util.forkProgram(Util.conj(opts, "server"));
    msg("Starting server...");
  }

  public void startFg() throws InvalidOptionsException {
    validateNotRemoteHost();

    Object[] args = {port,
                     Keyword.intern("pre-buffer"), preBuffer,
                     Keyword.intern("post-buffer"), postBuffer};

    if (useStockSoundfont) {
      args = Util.concat(args, new Object[]{Keyword.intern("stock"), true});
    }

    Util.callClojureFn("alda.server/start-server!", args);
  }

  // TODO: rewrite REPL as a client that communicates with a server
  public void startRepl() throws InvalidOptionsException {
    validateNotRemoteHost();

    Object[] args = {Keyword.intern("pre-buffer"), preBuffer,
                     Keyword.intern("post-buffer"), postBuffer};

    if (useStockSoundfont) {
      args = Util.concat(args, new Object[]{Keyword.intern("stock"), true});
    }

    Util.callClojureFn("alda.repl/start-repl!", args);
  }


  public void stop() {
    msg("Stopping server...");
  }

  public void restart()
    throws InvalidOptionsException, java.net.URISyntaxException, java.io.IOException {
    stop();
    System.out.println();
    startBg();
  }

}
