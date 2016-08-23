package alda;

public class AldaWorker extends AldaProcess {
  public AldaWorker(int port) {
    this.port = port;
  }

  public void upFg() throws InvalidOptionsException {
    Object[] args = {this.port};

    Util.callClojureFn("alda.worker/start-worker!", args);
  }
}
