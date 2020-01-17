plugins {
  id("org.jetbrains.kotlin.jvm") version("1.3.11")
  application
}

repositories {
  jcenter()
  mavenLocal()
}

dependencies {
  compile("org.jetbrains.kotlin:kotlin-stdlib-jdk8")
  testImplementation("org.jetbrains.kotlin:kotlin-test")
  testImplementation("org.jetbrains.kotlin:kotlin-test-junit")

  compile("com.illposed.osc:javaosc-core")
}

dependencies {
  constraints {
    compile("com.illposed.osc:javaosc-core:0.7-SNAPSHOT")
  }
}

application {
  mainClassName = "io.alda.player.MainKt"
}

val run by tasks.getting(JavaExec::class) {
  standardInput = System.`in`
}

val fatJar = task("fatJar", type = Jar::class) {
    baseName = "${project.name}-fat"
    manifest {
        attributes["Main-Class"] = "io.alda.player.MainKt"
    }
    from(configurations.compile.get().map({ if (it.isDirectory) it else zipTree(it) }))
    from(configurations.runtime.get().map({ if (it.isDirectory) it else zipTree(it) }))
    with(tasks["jar"] as CopySpec)
}

tasks {
    "build" {
        dependsOn(fatJar)
    }
}
