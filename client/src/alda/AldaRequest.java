package alda;

import com.google.gson.Gson;

import org.zeromq.ZContext;
import org.zeromq.ZMsg;
import org.zeromq.ZMQ;
import org.zeromq.ZMQ.PollItem;
import org.zeromq.ZMQ.Poller;
import org.zeromq.ZMQ.Socket;

public class AldaRequest {
  private final static int REQUEST_TIMEOUT = 2500; //  ms
  private final static int REQUEST_RETRIES = 3;    //  Before we abandon

  private transient String host;
  private transient int port;

  public AldaRequest(String host, int port) {
    this.host = host;
    this.port = port;
  }

  public String command;
  public String body;
  public AldaRequestOptions options;

  public String toJson() {
    Gson gson = new Gson();
    return gson.toJson(this);
  }

  private String sendRequest(String req, ZContext ctx, Socket client, int timeout, int retries)
    throws NoResponseException {
    if (retries < 0 || Thread.currentThread().isInterrupted()) {
      ctx.destroy();
      throw new NoResponseException("Alda server is down. To start the server, run `alda up`.");
    }

    assert (client != null);
    client.connect(this.host + ":" + this.port);

    ZMsg msg = new ZMsg();
    msg.addString(this.toJson());
    msg.addString(this.command);
    msg.send(client);

    PollItem items[] = {new PollItem(client, Poller.POLLIN)};
    int rc = ZMQ.poll(items, timeout);
    if (rc == -1) {
      throw new NoResponseException("Connection interrupted.");
    }

    if (items[0].isReadable()) {
      String response = client.recvStr();
      if (response == null) {
        throw new NoResponseException("Connection interrupted.");
      }
      return response;
    }

    // Old socket is confused; close it and open a new one
    ctx.destroySocket(client);
    client = ctx.createSocket(ZMQ.REQ);

    // Send request again, on new socket
    return sendRequest(req, ctx, client, timeout, retries - 1);
  }

  private String sendRequest(String req, ZContext ctx, Socket client, int timeout)
    throws NoResponseException {
    return sendRequest(req, ctx, client, timeout, REQUEST_RETRIES);
  }

  private String sendRequest(String req, ZContext ctx, Socket client)
    throws NoResponseException {
    return sendRequest(req, ctx, client, REQUEST_TIMEOUT, REQUEST_RETRIES);
  }

  public AldaResponse send(int timeout, int retries)
    throws NoResponseException {
    ZContext ctx = new ZContext();
    Socket client = ctx.createSocket(ZMQ.REQ);
    String res = sendRequest(this.toJson(), ctx, client, timeout, retries);
    return AldaResponse.fromJson(res);
  }

  public AldaResponse send(int timeout) throws NoResponseException {
    return send(timeout, REQUEST_RETRIES);
  }

  public AldaResponse send() throws NoResponseException {
    return send(REQUEST_TIMEOUT, REQUEST_RETRIES);
  }
}
