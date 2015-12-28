package alda;

import java.io.File;
import java.io.IOException;
import java.net.ConnectException;
import java.util.List;
import java.util.Map;
import java.net.UnknownHostException;
import java.net.URISyntaxException;
import java.util.concurrent.Callable;
import java.util.concurrent.TimeUnit;

import org.fusesource.jansi.AnsiConsole;
import static org.fusesource.jansi.Ansi.*;
import static org.fusesource.jansi.Ansi.Color.*;

import clojure.lang.Keyword;

import static us.bpsm.edn.Keyword.newKeyword;
import us.bpsm.edn.parser.Parseable;
import us.bpsm.edn.parser.Parser;
import us.bpsm.edn.parser.Parsers;

import org.apache.http.HttpEntity;
import org.apache.http.HttpResponse;
import org.apache.http.client.config.RequestConfig;
import org.apache.http.client.HttpResponseException;
import org.apache.http.client.methods.CloseableHttpResponse;
import org.apache.http.client.methods.HttpGet;
import org.apache.http.client.methods.HttpDelete;
import org.apache.http.client.methods.HttpPost;
import org.apache.http.client.methods.HttpPut;
import org.apache.http.client.methods.HttpRequestBase;
import org.apache.http.client.ResponseHandler;
import org.apache.http.conn.ConnectTimeoutException;
import org.apache.http.entity.ContentType;
import org.apache.http.entity.FileEntity;
import org.apache.http.entity.StringEntity;
import org.apache.http.impl.client.CloseableHttpClient;
import org.apache.http.impl.client.HttpClientBuilder;
import org.apache.http.impl.client.HttpClients;
import org.apache.http.util.EntityUtils;

import net.jodah.recurrent.Recurrent;
import net.jodah.recurrent.RetryPolicy;

public class AldaServer {
  private String host;
  private int port;
  private int preBuffer;
  private int postBuffer;
  private CloseableHttpClient httpclient;

  public String getHost() { return host; }
  public int getPort() { return port; }

  public AldaServer(String host, int port, int preBuffer, int postBuffer) {
    this.host = normalizeHost(host);
    this.port = port;
    this.preBuffer = preBuffer;
    this.postBuffer = postBuffer;

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

  private String doRequest(HttpRequestBase httpRequest) throws IOException {
    try {
      ResponseHandler<String> responseHandler = new ResponseHandler<String>() {
        @Override
        public String handleResponse(final HttpResponse response)
          throws HttpResponseException, IOException {
          int status = response.getStatusLine().getStatusCode();
          HttpEntity entity = response.getEntity();
          String responseBody = entity != null ? EntityUtils.toString(entity) : null;

          if (response.getFirstHeader("X-Alda-Version") == null) {
            throw new HttpResponseException(status, "Missing X-Alda-Version header. " +
                                                    "Probably not an Alda server.");
          } else if (status < 200 || status > 299) {
            throw new HttpResponseException(status, responseBody);
          } else {
            return responseBody;
          }
        }
      };

      String responseBody = httpclient.execute(httpRequest, responseHandler);
      return responseBody;
    } finally {
      httpclient.close();
    }
  }

  private String getRequest(String endpoint) throws IOException {
    HttpGet httpget = new HttpGet(host + ":" + port + endpoint);
    return doRequest(httpget);
  }

  private String deleteRequest(String endpoint) throws IOException {
    HttpDelete httpdelete = new HttpDelete(host + ":" + port + endpoint);
    return doRequest(httpdelete);
  }

  private String postRequest(String endpoint, HttpEntity entity)
    throws IOException {
    HttpPost httppost = new HttpPost(host + ":" + port + endpoint);
    httppost.setEntity(entity);
    return doRequest(httppost);
  }

  private String postString(String endpoint, String payload)
    throws IOException {
    StringEntity entity = new StringEntity(payload);
    return postRequest(endpoint, entity);
  }

  private String postFile(String endpoint, File payload)
    throws IOException {
    FileEntity entity = new FileEntity(payload);
    return postRequest(endpoint, entity);
  }

  private String putRequest(String endpoint, HttpEntity entity)
    throws IOException {
    HttpPut httpput = new HttpPut(host + ":" + port + endpoint);
    httpput.setEntity(entity);
    return doRequest(httpput);
  }

  private String putString(String endpoint, String payload)
    throws IOException {
    StringEntity entity = new StringEntity(payload);
    return putRequest(endpoint, entity);
  }

  private String putFile(String endpoint, File payload)
    throws IOException {
    FileEntity entity = new FileEntity(payload);
    return putRequest(endpoint, entity);
  }

  private boolean checkForConnection() throws Exception {
    try {
      getRequest("/");
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

  private void startServerIfNeeded() throws Exception {
    try {
      assertServerUp();
    } catch (Exception e) {
      startBg();
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
          getRequest("/");
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
          getRequest("/");
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

    Util.callClojureFn("alda.server/start-server!", args);
  }

  // TODO: rewrite REPL as a client that communicates with a server
  public void startRepl() throws InvalidOptionsException {
    assertNotRemoteHost();

    Object[] args = {Keyword.intern("pre-buffer"), preBuffer,
                     Keyword.intern("post-buffer"), postBuffer};

    Util.callClojureFn("alda.repl/start-repl!", args);
  }

  public void stop() throws Exception {
    boolean serverAlreadyDown = !checkForConnection();
    if (serverAlreadyDown) {
      msg("Server already down.");
      return;
    }

    msg("Stopping Alda server...");
    getRequest("/stop");

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
    msg(getRequest("/version"));
  }

  public void info() throws Exception {
    assertServerUp();
    String ednString = getRequest("/");
    Parseable pbr = Parsers.newParseable(ednString);
    Parser p = Parsers.newParser(Parsers.defaultConfiguration());
    Map<?, ?> m = (Map<?, ?>) p.nextValue(pbr);

    String status      = (String) m.get(newKeyword("status"));
    String version     = (String) m.get(newKeyword("version"));
    String filename    = (String) m.get(newKeyword("filename"));
    Boolean isModified = (Boolean) m.get(newKeyword("modified?"));
    Long lineCount     = (Long) m.get(newKeyword("line-count"));
    Long charCount     = (Long) m.get(newKeyword("char-count"));
    @SuppressWarnings("unchecked") List<Map<?, ?>> instruments =
      (List<Map<?, ?>>) m.get(newKeyword("instruments"));

    msg("Server status: " + status);
    msg("Server version: " + version);
    msg("Filename: " + (filename != null ? filename : "(new score)"));
    msg("Modified: " + (isModified ? "yes" : "no"));
    msg("Line count: " + lineCount);
    msg("Character count: " + charCount);
    if (instruments.isEmpty()) {
      msg("Instruments: (none)");
    } else {
      msg("Instruments:");
      for (Map<?, ?> instrument : instruments) {
        String instrumentName  = (String) instrument.get(newKeyword("name"));
        String instrumentStock = (String) instrument.get(newKeyword("stock"));
        String instrumentString = instrumentStock + " (" + instrumentName + ")";
        msg("  \u2022 " + instrumentString);
      }
    }
  }

  public void score(String mode) throws Exception {
    assertServerUp();
    System.out.println(getRequest("/score/" + mode));
  }

  public void delete() throws Exception {
    assertServerUp();
    deleteRequest("/score");
    msg("New score initialized.");
  }

  public void play() throws Exception {
    assertServerUp();
    getRequest("/play");
    msg("Playing score...");
  }

  public void play(String code, boolean replaceScore) throws Exception {
    startServerIfNeeded();
    String result = replaceScore ? putString("/play", code)
                                 : postString("/play", code);
    msg("Playing code...");
  }

  public void play(File file, boolean replaceScore) throws Exception {
    startServerIfNeeded();
    String result = replaceScore ? putFile("/play", file)
                                 : postFile("/play", file);
    msg("Playing file...");
  }

  public void parse(String code, String mode) throws Exception {
    startServerIfNeeded();
    System.out.println(postString("/parse/" + mode, code));
  }

  public void parse(File file, String mode) throws Exception {
    startServerIfNeeded();
    System.out.println(postFile("/parse/" + mode, file));
  }

  public void append(String code) throws Exception {
    startServerIfNeeded();
    postString("/add", code);
    msg("Appended code to score.");
  }

  public void append(File file) throws Exception {
    startServerIfNeeded();
    postFile("/add", file);
    msg("Appended file to score.");
  }

}
