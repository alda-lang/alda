plugins {
  id("org.jetbrains.kotlin.jvm") version("1.3.60")
  application
}

repositories {
  mavenLocal()
  mavenCentral()
  jcenter()
}

dependencies {
  implementation(kotlin("stdlib-jdk8"))
  implementation("com.beust:klaxon:5.0.1")
  implementation("com.github.ajalt:clikt:2.4.0")

  // implementation("com.illposed.osc:javaosc-core:0.7")
  implementation("io.djy:javaosc:0.7") {
    exclude("org.slf4j", "slf4j-api")
    exclude("org.slf4j", "slf4j-ext")
    exclude("org.slf4j", "slf4j-log4j12")
    exclude("log4j", "log4j")
  }

  implementation("io.github.soc:directories:11")

  // logging
  implementation("io.github.microutils:kotlin-logging:1.7.7")
  implementation("org.slf4j:slf4j-api:1.7.30")
  implementation("org.apache.logging.log4j:log4j-slf4j-impl:2.13.0")
  implementation("org.apache.logging.log4j:log4j-api:2.13.0")
  implementation("org.apache.logging.log4j:log4j-core:2.13.0")

  testImplementation(kotlin("test"))
  testImplementation(kotlin("test-junit"))
}

application {
  mainClassName = "io.alda.player.MainKt"
}

val run by tasks.getting(JavaExec::class) {
  standardInput = System.`in`
}

val fatJar = task("fatJar", type = Jar::class) {
  archiveBaseName.set("${project.name}-fat")
  duplicatesStrategy = DuplicatesStrategy.INCLUDE
  manifest {
    attributes["Main-Class"] = "io.alda.player.MainKt"
    attributes["Multi-Release"] = "true"
  }
  from(configurations.compileClasspath.get().map({ if (it.isDirectory) it else zipTree(it) }))
  from(configurations.runtimeClasspath.get().map({ if (it.isDirectory) it else zipTree(it) }))
  with(tasks["jar"] as CopySpec)
}

tasks {
  "build" {
    dependsOn(fatJar)
  }
}
