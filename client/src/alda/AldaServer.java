package alda;

import org.fusesource.jansi.AnsiConsole;
import static org.fusesource.jansi.Ansi.*;
import static org.fusesource.jansi.Ansi.Color.*;

public class AldaServer {
  private String host;
  private int port;

  public String getHost() { return host; }
  public int getPort() { return port; }

  public AldaServer(String host, int port) {
    this.host = normalizeHost(host);
    this.port = port;
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

  public void start() throws InvalidOptionsException {
    String hostWithoutProtocol = host.replaceAll("https?://", "");

    if (!hostWithoutProtocol.equals("localhost")) {
      throw new InvalidOptionsException("Alda servers cannot be started " +
                                        "remotely.");
    }

    msg("Starting server...");
  }

  public void stop() {
    msg("Stopping server...");
  }

  public void restart() throws InvalidOptionsException {
    start();
    System.out.println();
    stop();
  }

}
