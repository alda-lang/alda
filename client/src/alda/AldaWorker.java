package alda;

public class AldaWorker extends AldaProcess {
  public AldaWorker(int port, boolean verbose) {
    this.port = port;
    this.verbose = verbose;
  }

  public void upFg() throws InvalidOptionsException {
    Object[] args = {this.port, this.verbose};

    Util.callClojureFn("alda.worker/start-worker!", args);
  }
}
