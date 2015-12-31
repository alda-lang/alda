package alda;

import java.util.List;
import java.util.Map;

public class AldaServerInfo {
  public String status;
  public String version;
  public String filename;
  public Boolean isModified;
  public Long lineCount;
  public Long charCount;
  public List<Map<?,?>> instruments;

  public AldaServerInfo(String status, String version, String filename,
                        Boolean isModified, Long lineCount, Long charCount,
                        List<Map<?,?>> instruments) {
    this.status = status;
    this.version = version;
    this.filename = filename;
    this.isModified = isModified;
    this.lineCount = lineCount;
    this.charCount = charCount;
    this.instruments = instruments;
  }
}
