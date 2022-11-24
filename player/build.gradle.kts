plugins {
  id("org.jetbrains.kotlin.jvm") version("1.7.21")
  application
}

java {
  targetCompatibility = JavaVersion.VERSION_1_8
}

repositories {
  mavenLocal()
  mavenCentral()
}

dependencies {
  implementation(kotlin("stdlib-jdk8"))
  implementation(kotlin("reflect"))
  implementation("com.beust:klaxon:5.6")
  implementation("com.github.ajalt:clikt:2.4.0")

  implementation("com.illposed.osc:javaosc-core:0.8") {
    exclude("org.slf4j", "slf4j-api")
    exclude("org.slf4j", "slf4j-ext")
    exclude("org.slf4j", "slf4j-log4j12")
    exclude("log4j", "log4j")
  }

  implementation("io.github.soc:directories:11")

  // logging
  implementation("io.github.microutils:kotlin-logging:1.7.7")
  implementation("org.slf4j:slf4j-api:1.7.30")
  implementation("org.apache.logging.log4j:log4j-slf4j-impl:2.17.0")
  implementation("org.apache.logging.log4j:log4j-api:2.17.0")
  implementation("org.apache.logging.log4j:log4j-core:2.17.0")

  testImplementation(kotlin("test"))
  testImplementation(kotlin("test-junit"))
}

application {
  mainClass.set("io.alda.player.MainKt")
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
