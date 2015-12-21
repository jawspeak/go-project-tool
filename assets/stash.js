var render = function() {

  var authorStats = function(data) {
    var byAuthor = d3.map()
    data.PullRequests.forEach(function (e) {
      if (!byAuthor.has(e.author_ldap)) {
        byAuthor.set(e.author_ldap, []);
      }
      byAuthor.get(e.author_ldap).push(e);
    });
    //byAuthor.forEach(function(k,v) {
    //  v.sort(function(a,b){
    //    return a.created_datetime - b.created_datetime;
    //  });
    //});
    return byAuthor;
  };

  var renderAuthorHeader = function(author, prs) {
    var personContainer = d3.selectAll(".teamlisting")
        .append("div")
        .attr("class", "person");
    personContainer.append("h4").text(author);
    personContainer.append("ul")
        .append("li").text("" + prs.length + " PRs") // TODO (N Merged)
        // .append("li").text("Commits")
        .append("li").text("Commented in __ other people's PR's")
    return personContainer;
  };

  var renderPrs = function(byAuthor, personContainer, prs) {
    var width = 150,
        height = 75;
    prs.forEach(function(onePr) {
      /* Creates 2 fixed columns of nodes:
       PR       comment authors
       |      /----- A1
       o------------ A2
       |      \----- A3

       The comment authors are sorted by # of comments. */
      var nodes = [];
      var prNode = { title: onePr.title, type: "PR", root: onePr};
      var links = [];
      for (var author in onePr.comments_by_author_ldap) {
        var count = onePr.comments_by_author_ldap[author];
        var is_approval = onePr.approvals_by_author_ldap[author] > 0;
        var authorNode = {
          title: (count > 1 ? "" + count + " comments" : "" + count + " comment") + " by " + author +
          (is_approval ? " (approval)" : ""),
          type: "comments", author: author, is_approval: is_approval, count: count
        };
        nodes.push(authorNode);
        links.push({source: authorNode, target: prNode, value: authorNode.count});
      }
      nodes.sort(function(a,b) {
        return a.count - b.count;
      });
      nodes.unshift(prNode);

      //console.log(nodes);
      //console.log(links);

      var force = d3.layout.force()
          .charge(-100)
          .linkDistance(20)
          .size([width, height]);

      force.nodes(nodes)
          .links(links)
          .start();

      var svg = personContainer
          .append("svg")
          .attr("width", width)
          .attr("height", height);

      var link = svg.selectAll(".link")
          .data(links)
          .enter().append("line")
          .attr("class", "link")
          .style("stroke-width", function(d) { return 1.5 * d.value + "px";});

      var node = svg.selectAll(".node")
          .data(nodes)
          .enter().append("circle")
          .attr("class", "node")
          .attr("r", 5)
          .style("fill", function(n) {
            if (n.type == "PR") {
              return "orange";
            }
            return "blue";
          });

      node.attr("cx", function(d) {
        switch (d.type) {
          case "PR":
            return 10;
          case "comments":
            return 100;
          default:
            throw "error in d.type";
        }
      });

      node.append("title")
          .text(function(d) {
            return d.title; })

      force.on("tick", function() {
        link.attr("x1", function(d) { return d.source.x; })
            .attr("y1", function(d) { return d.source.y; })
            .attr("x2", function(d) { return d.target.x; })
            .attr("y2", function(d) { return d.target.y; });
        node.attr("cx", function(d) { return d.x; })
            .attr("cy", function(d) { return d.y; });
      });
    })
  }

  var renderForcePRPlot = function(data) {
    var byAuthor = authorStats(data);
    byAuthor.forEach(function(author, prs) {
      var personContainer = renderAuthorHeader(author, prs);
      renderPrs(byAuthor, personContainer, prs);
    });

  }

  d3.json('/assets/pr-data-pull.json', function(data) {
    renderForcePRPlot(data)
  });
};

render();


/*

 */
