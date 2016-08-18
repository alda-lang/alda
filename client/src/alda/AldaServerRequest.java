package alda;

import com.google.gson.Gson;

import org.jeromq.ZContext;
import org.jeromq.ZMQ;
import org.jeromq.ZMQ.PollItem;
import org.jeromq.ZMQ.Poller;
import org.jeromq.ZMQ.Socket;

public class AldaServerRequest {
  private final static int REQUEST_TIMEOUT = 2500; //  ms
  private final static int REQUEST_RETRIES = 3;    //  Before we abandon

  private String host;
  private int port;

  public AldaServerRequest(String host, int port) {
    this.host = host;
    this.port = port;
  }

  public String command;
  public String body;
  public AldaServerRequestOptions options;
  public boolean confirming;

  public String toJson() {
    Gson gson = new Gson();
    return gson.toJson(this);
  }

  private String sendRequest(String req, ZContext ctx, Socket client, int timeout, int retries)
    throws ServerResponseException {
    if (retries <= 0 || Thread.currentThread().isInterrupted()) {
      ctx.destroy();
      throw new ServerResponseException("Alda server is down. To start the server, run `alda up`.");
    }

    assert (client != null);
    client.connect(this.host + ":" + this.port);
    client.send(req);

    PollItem items[] = {new PollItem(client, Poller.POLLIN)};
    int rc = ZMQ.poll(items, timeout);
    if (rc == -1) {
      throw new ServerResponseException("Connection interrupted.");
    }

    if (items[0].isReadable()) {
      String response = client.recvStr();
      if (response == null) {
        throw new ServerResponseException("Connection interrupted.");
      }
      return response;
    }

    // Old socket is confused; close it and open a new one
    ctx.destroySocket(client);
    client = ctx.createSocket(ZMQ.REQ);

    // Send request again, on new socket
    return sendRequest(req, ctx, client, retries - 1, timeout);
  }

  private String sendRequest(String req, ZContext ctx, Socket client, int timeout)
    throws ServerResponseException {
    return sendRequest(req, ctx, client, timeout, REQUEST_RETRIES);
  }

  private String sendRequest(String req, ZContext ctx, Socket client)
    throws ServerResponseException {
    return sendRequest(req, ctx, client, REQUEST_TIMEOUT, REQUEST_RETRIES);
  }

  public AldaServerResponse send(int timeout, int retries)
    throws ServerResponseException {
    ZContext ctx = new ZContext();
    Socket client = ctx.createSocket(ZMQ.REQ);
    String res = sendRequest(this.toJson(), ctx, client, retries, timeout);
    return AldaServerResponse.fromJson(res);
  }

  public AldaServerResponse send(int timeout) throws ServerResponseException {
    return send(timeout, REQUEST_RETRIES);
  }

  public AldaServerResponse send() throws ServerResponseException {
    return send(REQUEST_TIMEOUT, REQUEST_RETRIES);
  }
}
