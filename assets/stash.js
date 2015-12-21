var render = function() {

  var authorStats = function(data) {
    var byAuthor = d3.map()
    data.pull_requests.forEach(function (pr) {
      if (!byAuthor.has(pr.author_ldap)) {
        byAuthor.set(pr.author_ldap, {
          prs: [],
          total_comments_left: 0,
          others_prs_commented_in: 0,
          others_prs_approved: 0,
          full_name: pr.author_fullname});
      }
      byAuthor.get(pr.author_ldap).prs.push(pr);
      console.log(pr);
    });
    data.pull_requests.forEach(function(pr) {
      for (commenter in pr.comments_by_author_ldap) {
        if (byAuthor.get(commenter)) { // if someone we care about
          byAuthor.get(commenter).total_comments_left += pr.comments_by_author_ldap[commenter];
          if (pr.author_ldap != commenter) {
            byAuthor.get(commenter).others_prs_commented_in += 1;
          }
        }
      }
      for (approver in pr.approvals_by_author_ldap) {
        if (byAuthor.get(approver) && pr.author_ldap != approver) {
          byAuthor.get(approver).others_prs_approved += 1;
        }
      }
    });

    byAuthor.forEach(function(k,v) {
      v.prs.sort(function(a,b){
        return b.created_datetime - a.created_datetime;
      });
    });
    return byAuthor;
  };

  var renderAuthorHeader = function(author, stats) {
    var personContainer = d3.selectAll(".teamlisting")
        .append("div")
        .attr("class", "person");
    personContainer.append("h4").text(stats.full_name);

    var merged = stats.prs.filter(function(pr) { return pr.state === "MERGED";}).length;

    var ul = personContainer.append("ul");
    ul.append("li").text("" + stats.prs.length + " PRs").append("small").text(" (" + merged + " merged)");
    ul.append("li").text("Commented in " + stats.others_prs_commented_in + " other people's PRs ")
        .append("small").text("(Approved " + stats.others_prs_approved + " PR's, left "
        + stats.total_comments_left + " total comments)");
    return personContainer;
  };

  var renderPrs = function(byAuthor, personContainer, stats) {
    var width = 150,
        height = 75;

    stats.prs.forEach(function(onePr) {
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

      var prDiv = personContainer
          .append("div")
          .attr("class", onePr.state === "OPEN" ? "pr-container unmerged" : "pr-container")
          .append("a")
          .attr("href", onePr.self_url);
      prDiv.append("div").append("small").text(onePr.title);

      var svg = prDiv
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
    byAuthor.forEach(function(author, stats) {
      var personContainer = renderAuthorHeader(author, stats);
      renderPrs(byAuthor, personContainer, stats);
    });

  }

  d3.json('/assets/pr-data-pull.json', function(data) {
    renderForcePRPlot(data)
  });
};

render();
