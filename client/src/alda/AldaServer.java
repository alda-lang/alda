package alda;

import java.io.IOException;
import java.net.ConnectException;
import java.net.UnknownHostException;
import java.net.URISyntaxException;

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
    throws InvalidOptionsException, URISyntaxException, IOException {
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
    throws InvalidOptionsException, URISyntaxException, IOException {
    stop();
    System.out.println();
    startBg();
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

  private void serverUp() {
    msg(ansi().a("Server up ").fg(GREEN).a("\u2713").reset().toString());
  }

  private void serverDown() {
    msg(ansi().a("Server down ").fg(RED).a("\u2717").reset().toString());
  }

  public void status() throws IOException {
    try {
      get("/");
      serverUp();
    } catch (HttpResponseException e) {
      serverDown();
    } catch (ConnectException e) {
      serverDown();
    } catch (ConnectTimeoutException e) {
      serverDown();
    } catch (UnknownHostException e) {
      error("Invalid hostname. Please check to make sure it is correct.");
    }
  }

}
