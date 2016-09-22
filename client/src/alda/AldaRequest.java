package alda;

import com.google.gson.Gson;

import org.zeromq.ZContext;
import org.zeromq.ZMsg;
import org.zeromq.ZMQ;
import org.zeromq.ZMQ.PollItem;
import org.zeromq.ZMQ.Poller;
import org.zeromq.ZMQ.Socket;

public class AldaRequest {
  private static ZContext zContext = null;
  public static ZContext getZContext() {
    if (zContext == null) {
      zContext = new ZContext();
    }
    return zContext;
  }

  private static Socket dealerSocket = null;
  public static Socket getDealerSocket() {
    if (dealerSocket == null) {
      dealerSocket = getZContext().createSocket(ZMQ.DEALER);
    }
    return dealerSocket;
  }

  private final static int REQUEST_TIMEOUT = 2500; //  ms
  private final static int REQUEST_RETRIES = 3;    //  Before we abandon

  private transient String host;
  private transient int port;
  public transient byte[] workerToUse;

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

  private AldaResponse sendRequest(String req, ZContext ctx, Socket client, int timeout, int retries)
    throws NoResponseException {
    if (retries < 0 || Thread.currentThread().isInterrupted()) {
      throw new NoResponseException("Alda server is down. To start the server, run `alda up`.");
    }

    assert (client != null);
    client.connect(this.host + ":" + this.port);

    ZMsg msg = new ZMsg();
    msg.addString(this.toJson());
    if (this.workerToUse != null) {
      msg.add(this.workerToUse);
    }
    msg.addString(this.command);
    msg.send(client);

    PollItem items[] = {new PollItem(client, Poller.POLLIN)};
    int rc = ZMQ.poll(items, timeout);
    if (rc == -1) {
      throw new NoResponseException("Connection interrupted.");
    }

    if (items[0].isReadable()) {
      String address = client.recvStr();
      if (address == null) {
        throw new NoResponseException("Connection interrupted.");
      }

      String responseJson = client.recvStr();
      if (responseJson == null) {
        throw new NoResponseException("Connection interrupted.");
      }

      AldaResponse response = AldaResponse.fromJson(responseJson);

      byte[] workerAddress = client.recv(ZMQ.DONTWAIT);
      if (workerAddress != null) {
        response.workerAddress = workerAddress;
      }

      return response;
    }

    // Send request again until we're out of retries
    return sendRequest(req, ctx, client, timeout, retries - 1);
  }

  private AldaResponse sendRequest(String req, ZContext ctx, Socket client,
                                   int timeout)
    throws NoResponseException {
    return sendRequest(req, ctx, client, timeout, REQUEST_RETRIES);
  }

  private AldaResponse sendRequest(String req, ZContext ctx, Socket client)
    throws NoResponseException {
    return sendRequest(req, ctx, client, REQUEST_TIMEOUT, REQUEST_RETRIES);
  }

  public AldaResponse send(int timeout, int retries)
    throws NoResponseException {
    ZContext ctx = getZContext();
    Socket client = getDealerSocket();
    return sendRequest(this.toJson(), ctx, client, timeout, retries);
  }

  public AldaResponse send(int timeout) throws NoResponseException {
    return send(timeout, REQUEST_RETRIES);
  }

  public AldaResponse send() throws NoResponseException {
    return send(REQUEST_TIMEOUT, REQUEST_RETRIES);
  }
}
