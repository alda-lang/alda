package alda;

import java.io.IOException;
import java.net.ConnectException;
import java.net.UnknownHostException;
import java.net.URISyntaxException;

import java.util.concurrent.Callable;
import java.util.concurrent.TimeUnit;

import org.fusesource.jansi.AnsiConsole;
import static org.fusesource.jansi.Ansi.*;
import static org.fusesource.jansi.Ansi.Color.*;

import clojure.lang.Keyword;

import org.apache.http.HttpEntity;
import org.apache.http.HttpResponse;
import org.apache.http.client.config.RequestConfig;
import org.apache.http.client.HttpResponseException;
import org.apache.http.client.methods.HttpGet;
import org.apache.http.client.ResponseHandler;
import org.apache.http.conn.ConnectTimeoutException;
import org.apache.http.impl.client.CloseableHttpClient;
import org.apache.http.impl.client.HttpClientBuilder;
import org.apache.http.util.EntityUtils;

import net.jodah.recurrent.Recurrent;
import net.jodah.recurrent.RetryPolicy;

public class AldaServer {
  private String host;
  private int port;
  private int preBuffer;
  private int postBuffer;
  private boolean useStockSoundfont;
  private CloseableHttpClient httpclient;

  public String getHost() { return host; }
  public int getPort() { return port; }

  public AldaServer(String host, int port, int preBuffer, int postBuffer,
                    boolean useStockSoundfont) {
    this.host = normalizeHost(host);
    this.port = port;
    this.preBuffer = preBuffer;
    this.postBuffer = postBuffer;
    this.useStockSoundfont = useStockSoundfont;

    RequestConfig config = RequestConfig.custom()
                                        .setConnectTimeout(5000)
                                        .setConnectionRequestTimeout(5000)
                                        .setSocketTimeout(5000)
                                        .build();

    this.httpclient = HttpClientBuilder.create()
                                       .setDefaultRequestConfig(config)
                                       .setConnectionManagerShared(true)
                                       .disableAutomaticRetries()
                                       .build();

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

  private void assertNotRemoteHost() throws InvalidOptionsException {
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

  private String get(String endpoint) throws IOException {
    try {
      HttpGet httpget = new HttpGet(host + ":" + port + endpoint);

      ResponseHandler<String> responseHandler = new ResponseHandler<String>() {
        @Override
        public String handleResponse(final HttpResponse response)
          throws HttpResponseException, IOException {
          int status = response.getStatusLine().getStatusCode();
          if (status < 200 || status > 299) {
            throw new HttpResponseException(status, "Unexpected response status: " + status);
          } else if (response.getFirstHeader("X-Alda-Version") == null) {
            throw new HttpResponseException(status, "Missing X-Alda-Version header. " +
                                                    "Probably not an Alda server.");
          } else {
            HttpEntity entity = response.getEntity();
            return entity != null ? EntityUtils.toString(entity) : null;
          }
        }
      };

      String responseBody = httpclient.execute(httpget, responseHandler);
      return responseBody;
    } finally {
      httpclient.close();
    }
  }

  private boolean checkForConnection() throws Exception {
    try {
      get("/");
      return true;
    } catch (UnknownHostException e) {
      throw new Exception("Invalid hostname. " +
                          "Please check to make sure it is correct.");
    } catch (Exception e) {
      return false;
    }
  }

  private void assertServerUp() throws Exception {
    boolean serverUp = checkForConnection();
    if (!serverUp) {
      throw new Exception("The Alda server is down.");
    }
  }

  // Keeps trying to connect to the server for 30 seconds.
  // Returns true if/when it gets a successful response.
  // Returns false if it doesn't get one within 30 seconds.
  private boolean waitForConnection() {
    RetryPolicy retryPolicy = new RetryPolicy()
      .withDelay(500, TimeUnit.MILLISECONDS)
      .withMaxDuration(30, TimeUnit.SECONDS)
      .retryFor(null);

    Callable<Boolean> ping = new Callable<Boolean>() {
      public Boolean call() throws ConnectException {
        try {
          get("/");
          return new Boolean(true);
        } catch (ConnectException e) {
          return null;
        } catch (Exception e) {
          return new Boolean(false);
        }
      }
    };

    Boolean serverUp = Recurrent.get(ping, retryPolicy);
    return serverUp == null ? false : serverUp.booleanValue();
  }

  // Keeps trying to connect to the server for 30 seconds.
  // Returns true as soon as it does NOT get a successful response.
  // Returns false if it's been 30 seconds and it's still getting a response.
  private boolean waitForLackOfConnection() {
    RetryPolicy retryPolicy = new RetryPolicy()
      .withDelay(500, TimeUnit.MILLISECONDS)
      .withMaxDuration(30, TimeUnit.SECONDS)
      .retryFor(null);

    Callable<Boolean> ping = new Callable<Boolean>() {
      public Boolean call() {
        try {
          get("/");
          return null;
        } catch (Exception e) {
          return new Boolean(true);
        }
      }
    };

    Boolean serverDown = Recurrent.get(ping, retryPolicy);
    return serverDown == null ? false : serverDown.booleanValue();
  }

  public void startBg() throws Exception {
    assertNotRemoteHost();

    boolean serverAlreadyUp = checkForConnection();
    if (serverAlreadyUp) {
      msg("Server already up.");
      return;
    }

    Object[] opts = {"--host", host, "--port", Integer.toString(port),
                     "--pre-buffer", Integer.toString(preBuffer),
                     "--post-buffer", Integer.toString(postBuffer)};

    if (useStockSoundfont) {
      opts = Util.conj(opts, "--stock");
    }

    Util.forkProgram(Util.conj(opts, "server"));
    msg("Starting Alda server...");

    boolean serverUp = waitForConnection();
    if (serverUp) {
      serverUp();
    } else {
      serverDown();
    }
  }

  public void startFg() throws InvalidOptionsException {
    assertNotRemoteHost();

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
    assertNotRemoteHost();

    Object[] args = {Keyword.intern("pre-buffer"), preBuffer,
                     Keyword.intern("post-buffer"), postBuffer};

    if (useStockSoundfont) {
      args = Util.concat(args, new Object[]{Keyword.intern("stock"), true});
    }

    Util.callClojureFn("alda.repl/start-repl!", args);
  }

  public void stop() throws Exception {
    boolean serverAlreadyDown = !checkForConnection();
    if (serverAlreadyDown) {
      msg("Server already down.");
      return;
    }

    msg("Stopping Alda server...");
    get("/stop");

    boolean serverIsDown = waitForLackOfConnection();
    if (serverIsDown) {
      serverDown(true);
    } else {
      throw new Exception("Failed to stop server.");
    }
  }

  public void restart() throws Exception {
    stop();
    System.out.println();
    startBg();
  }

  public void status() throws Exception {
    boolean serverIsUp = checkForConnection();
    if (serverIsUp) {
      serverUp();
    } else {
      serverDown();
    }
  }

  public void version() throws Exception {
    assertServerUp();
    msg(get("/version"));
  }

}
