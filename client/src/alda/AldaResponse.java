package alda;

import com.google.gson.Gson;

public class AldaResponse {
  public boolean success;
  public boolean pending;
  public String signal;
  public String body;
  public byte[] workerAddress;
  public boolean noWorker;

  public static AldaResponse fromJson(String json) {
    Gson gson = new Gson();
    return gson.fromJson(json, AldaResponse.class);
  }
}
