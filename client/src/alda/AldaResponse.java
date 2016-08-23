package alda;

import com.google.gson.Gson;

public class AldaResponse {
  public boolean success;
  public String signal;
  public String body;

  public static AldaResponse fromJson(String json) {
    Gson gson = new Gson();
    return gson.fromJson(json, AldaResponse.class);
  }
}
