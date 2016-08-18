package alda;

import com.google.gson.Gson;

public class AldaServerResponse {
  public boolean success;
  public String signal;
  public String body;

  public static AldaServerResponse fromJson(String json) {
    Gson gson = new Gson();
    return gson.fromJson(json, AldaServerResponse.class);
  }
}
